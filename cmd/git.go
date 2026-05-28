package main

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// setupGitConfig sets up the git configuration. It does NOT embed the GitHub
// token into the remote URL; the token is supplied per-push via
// http.extraheader (see pushReadme) so it never persists in .git/config.
func setupGitConfig(owner, name, email string, hideRepoInfo bool) error {
	if name == "" {
		name = "GitHub Action"
	}

	if email == "" {
		email = "action@github.com"
	}

	if err := runGitCommand(hideRepoInfo, "config", "--global", "--add", "safe.directory", "/github/workspace"); err != nil {
		return fmt.Errorf("git config safe.directory error: %v", sanitizeError(err, "", owner))
	}

	if err := runGitCommand(hideRepoInfo, "config", "--global", "user.name", name); err != nil {
		return fmt.Errorf("git config user.name error: %v", sanitizeError(err, "", owner))
	}

	if err := runGitCommand(hideRepoInfo, "config", "--global", "user.email", email); err != nil {
		return fmt.Errorf("git config user.email error: %v", sanitizeError(err, "", owner))
	}

	remoteURL := fmt.Sprintf("https://github.com/%s/%s.git", owner, owner)
	if err := runGitCommand(hideRepoInfo, "remote", "set-url", "origin", remoteURL); err != nil {
		return fmt.Errorf("git remote set-url error: %v", sanitizeError(err, "", owner))
	}

	return nil
}

// hasReadmeChanged checks if README.md has changed
func hasReadmeChanged() (bool, error) {
	cmd := exec.Command("git", "status", "--porcelain", "README.md")
	output, err := cmd.Output()
	if err != nil {
		return false, err
	}

	return strings.TrimSpace(string(output)) != "", nil
}

// commitAndPushReadme Commit and push changes if README.md has changed.
// The token is supplied to the push via GIT_CONFIG_* env vars and never
// appears on the command line or in .git/config.
func commitAndPushReadme(token, msg, branch string, hideRepoInfo bool) error {
	if branch == "" {
		branch = "main"
	}

	if msg == "" {
		msg = "📝 Update README.md"
	}

	if err := runGitCommand(hideRepoInfo, "add", "README.md"); err != nil {
		return err
	}

	if err := runGitCommand(hideRepoInfo, "commit", "-m", msg); err != nil {
		return err
	}

	return pushReadme(hideRepoInfo, token, branch)
}

// pushReadme performs `git push origin -- <branch>` while injecting an
// Authorization header via GIT_CONFIG_COUNT/KEY_0/VALUE_0 env vars. This
// matches the credential strategy used by actions/checkout and avoids
// writing the token into either argv or persisted git config.
func pushReadme(hideRepoInfo bool, token, branch string) error {
	// actions/checkout (default persist-credentials: true) writes its own
	// http.https://github.com/.extraheader into the workspace .git/config.
	// Git treats extraheader as multi-valued, so leaving it in place would
	// make git send two Authorization headers — GitHub rejects that with
	// "Duplicate header: Authorization" (HTTP 400) before validating either
	// token. Strip the persistent entry so only our per-push header is sent.
	if err := clearWorkspaceExtraheader(); err != nil {
		return err
	}

	authHeader := "Authorization: Basic " + base64.StdEncoding.EncodeToString([]byte("x-access-token:"+token))
	extraEnv := []string{
		"GIT_CONFIG_COUNT=1",
		"GIT_CONFIG_KEY_0=http.https://github.com/.extraheader",
		"GIT_CONFIG_VALUE_0=" + authHeader,
	}
	return runGitCommandEnv(hideRepoInfo, extraEnv, []string{authHeader, token}, "push", "origin", "--", branch)
}

// clearWorkspaceExtraheader removes any persisted
// http.https://github.com/.extraheader from the workspace .git/config so a
// pre-existing entry (typically left behind by actions/checkout) cannot
// collide with the per-push header injected via GIT_CONFIG_VALUE_0. We
// tolerate exit code 5 — git config's signal that the key is absent —
// because workflows without actions/checkout (or with persist-credentials:
// false) never set the entry in the first place.
func clearWorkspaceExtraheader() error {
	cmd := exec.Command("git", "config", "--local", "--unset-all", "http.https://github.com/.extraheader")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err == nil {
		return nil
	}
	if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 5 {
		return nil
	}
	return fmt.Errorf("failed to clear pre-existing extraheader: %v: %s", err, strings.TrimSpace(stderr.String()))
}

func runGitCommand(hideRepoInfo bool, args ...string) error {
	return runGitCommandEnv(hideRepoInfo, nil, nil, args...)
}

// runGitCommandEnv runs git with optional extra env vars and a list of
// secret strings to redact from any output or error message before it
// reaches the user. Callers pass redactions whenever args/env carry
// credentials so a malformed git error cannot leak them.
func runGitCommandEnv(hideRepoInfo bool, extraEnv, redactions []string, args ...string) error {
	cmd := exec.Command("git", args...)
	if len(extraEnv) > 0 {
		cmd.Env = append(os.Environ(), extraEnv...)
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	redact := func(s string) string {
		for _, r := range redactions {
			if r == "" {
				continue
			}
			s = strings.ReplaceAll(s, r, "[***]")
		}
		return s
	}

	if hideRepoInfo {
		// Suppress output when HIDE_REPO_INFO is true to prevent leaking sensitive information.
		// Even so, surface a sanitized version of stderr/stdout so that real failures
		// are debuggable instead of collapsing to a bare "exit status N".
		if err := cmd.Run(); err != nil {
			sanitizedErr := redact(err.Error())
			sanitizedStdout := sanitizeError(errors.New(redact(stdout.String())), "", "").Error()
			sanitizedStderr := sanitizeError(errors.New(redact(stderr.String())), "", "").Error()

			return fmt.Errorf("failed to run 'git %v': %v\n  stdout: %s\n  stderr: %s",
				args, sanitizedErr, sanitizedStdout, sanitizedStderr)
		}
	} else {
		if err := cmd.Run(); err != nil {
			writeSanitizedGitOutput(os.Stdout, redact(stdout.String()))
			writeSanitizedGitOutput(os.Stderr, redact(stderr.String()))

			return fmt.Errorf("failed to run 'git %v': %v", args, sanitizeError(errors.New(redact(err.Error())), "", ""))
		}

		writeSanitizedGitOutput(os.Stdout, redact(stdout.String()))
		writeSanitizedGitOutput(os.Stderr, redact(stderr.String()))
	}

	return nil
}

func writeSanitizedGitOutput(w *os.File, output string) {
	if output == "" {
		return
	}

	_, _ = fmt.Fprint(w, sanitizeError(errors.New(output), "", ""))
}

// sanitizeError removes sensitive information from error messages
func sanitizeError(err error, token, owner string) error {
	if err == nil {
		return nil
	}

	errMsg := err.Error()

	// Replace token with placeholder
	if token != "" {
		errMsg = strings.ReplaceAll(errMsg, token, "[***]")
	}

	gitHubTokenRegex := regexp.MustCompile(`\b(?:gh[opsu]_[A-Za-z0-9_]+|github_pat_[A-Za-z0-9_]+)\b`)
	errMsg = gitHubTokenRegex.ReplaceAllString(errMsg, "[***]")

	// Replace owner/username with placeholder
	if owner != "" {
		errMsg = strings.ReplaceAll(errMsg, owner, "[***]")
	}

	// Replace URLs with regex to redact entire URLs (http/https)
	// Matches: http(s)://anything until whitespace or end of string
	urlRegex := regexp.MustCompile(`https?://[^\s]+`)
	errMsg = urlRegex.ReplaceAllString(errMsg, "[***]")

	return fmt.Errorf("%s", errMsg)
}

// verifyCacheNotPushable returns an error when the cache file at cachePath
// is at risk of being committed to the repository — either because it is
// already tracked by git, or because it lives inside the repo without a
// matching .gitignore rule. Returns nil when the file is missing, gitignored,
// or located outside the repository (where git push cannot reach it).
func verifyCacheNotPushable(cachePath string) error {
	absCache, err := filepath.Abs(cachePath)
	if err != nil {
		return fmt.Errorf("failed to resolve cache file path: %v", err)
	}

	if _, err := os.Stat(absCache); os.IsNotExist(err) {
		return nil
	}

	cacheDir := filepath.Dir(absCache)
	rootCmd := exec.Command("git", "rev-parse", "--show-toplevel")
	rootCmd.Dir = cacheDir
	rootBytes, err := rootCmd.Output()
	if err != nil {
		// Cache file is not inside any git repository — git push cannot reach it.
		return nil
	}
	repoRoot := strings.TrimSpace(string(rootBytes))

	// Resolve symlinks on both sides before comparing — on macOS git returns
	// /private/var/... while filepath.Abs may yield /var/... for the same dir.
	resolvedRoot, err := filepath.EvalSymlinks(repoRoot)
	if err != nil {
		resolvedRoot = repoRoot
	}
	resolvedCache, err := filepath.EvalSymlinks(absCache)
	if err != nil {
		resolvedCache = absCache
	}

	rel, err := filepath.Rel(resolvedRoot, resolvedCache)
	if err != nil || strings.HasPrefix(rel, "..") {
		return nil
	}

	// Check 1: file already tracked? (catastrophic — gitignore won't help)
	tracked := exec.Command("git", "ls-files", "--error-unmatch", rel)
	tracked.Dir = repoRoot
	if err := tracked.Run(); err == nil {
		return fmt.Errorf(
			"cache file '%s' is already tracked by git — repo metadata may have leaked to history.\n"+
				"  To fix:\n"+
				"    1. git rm --cached %s\n"+
				"    2. Add '%s' to .gitignore\n"+
				"    3. Commit the fix",
			rel, rel, rel)
	}

	// Check 2: file gitignored? (-q: silent, exit code carries the answer)
	ignored := exec.Command("git", "check-ignore", "-q", rel)
	ignored.Dir = repoRoot
	err = ignored.Run()
	if err == nil {
		return nil
	}
	if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
		return fmt.Errorf(
			"cache file '%s' exists in the workspace but is not in .gitignore.\n"+
				"  Add this line to your repo's .gitignore:\n    %s",
			rel, rel)
	}
	return fmt.Errorf("failed to check gitignore status for '%s': %v", rel, err)
}
