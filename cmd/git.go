package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// setupGitConfig sets up the git configuration
func setupGitConfig(owner, token, name, email string, hideRepoInfo bool) error {
	if name == "" {
		name = "GitHub Action"
	}

	if email == "" {
		email = "action@github.com"
	}

	if err := runGitCommand(hideRepoInfo, "config", "--global", "--add", "safe.directory", "/github/workspace"); err != nil {
		return fmt.Errorf("git config safe.directory error: %v", sanitizeError(err, token, owner))
	}

	if err := runGitCommand(hideRepoInfo, "config", "--global", "user.name", name); err != nil {
		return fmt.Errorf("git config user.name error: %v", sanitizeError(err, token, owner))
	}

	if err := runGitCommand(hideRepoInfo, "config", "--global", "user.email", email); err != nil {
		return fmt.Errorf("git config user.email error: %v", sanitizeError(err, token, owner))
	}

	remoteURL := fmt.Sprintf("https://%s@github.com/%s/%s.git", token, owner, owner)
	if err := runGitCommand(hideRepoInfo, "remote", "set-url", "origin", remoteURL); err != nil {
		return fmt.Errorf("git remote set-url error: %v", sanitizeError(err, token, owner))
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

// commitAndPushReadme Commit and push changes if README.md has changed
func commitAndPushReadme(msg, branch string, hideRepoInfo bool) error {
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

	return runGitCommand(hideRepoInfo, "push", "origin", branch)
}

func runGitCommand(hideRepoInfo bool, args ...string) error {
	cmd := exec.Command("git", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if hideRepoInfo {
		// Suppress output when HIDE_REPO_INFO is true to prevent leaking sensitive information
		if err := cmd.Run(); err != nil {
			// Sanitize error output before returning
			sanitizedErr := err.Error()
			if stdout.Len() > 0 {
				sanitizedErr = strings.ReplaceAll(sanitizedErr, stdout.String(), "[output hidden]")
			}

			if stderr.Len() > 0 {
				sanitizedErr = strings.ReplaceAll(sanitizedErr, stderr.String(), "[output hidden]")
			}

			return fmt.Errorf("failed to run git command: %v", sanitizedErr)
		}
	} else {
		if err := cmd.Run(); err != nil {
			writeSanitizedGitOutput(os.Stdout, stdout.String())
			writeSanitizedGitOutput(os.Stderr, stderr.String())

			return fmt.Errorf("failed to run 'git %v': %v", args, sanitizeError(err, "", ""))
		}

		writeSanitizedGitOutput(os.Stdout, stdout.String())
		writeSanitizedGitOutput(os.Stderr, stderr.String())
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
