package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

// setupGitConfig sets up the git configuration
func setupGitConfig(owner, token, name, email string) error {
	hideRepoInfo := os.Getenv("HIDE_REPO_INFO") == "true"

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
func commitAndPushReadme(msg, branch string) error {
	hideRepoInfo := os.Getenv("HIDE_REPO_INFO") == "true"

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

	if hideRepoInfo {
		// Suppress output when HIDE_REPO_INFO is true to prevent leaking sensitive information
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			// Sanitize error output before returning
			sanitizedErr := strings.ReplaceAll(err.Error(), stdout.String(), "[output hidden]")
			sanitizedErr = strings.ReplaceAll(sanitizedErr, stderr.String(), "[output hidden]")
			return fmt.Errorf("failed to run git command: %v", sanitizedErr)
		}
	} else {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to run 'git %v': %v", args, err)
		}
	}

	return nil
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
