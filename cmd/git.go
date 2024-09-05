package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// setupGitConfig sets up the git configuration
func setupGitConfig(owner, token, name, email string) error {
	if name == "" {
		name = "GitHub Action"
	}

	if email == "" {
		email = "action@github.com"
	}

	if err := runGitCommand("config", "--global", "--add", "safe.directory", "/github/workspace"); err != nil {
		return fmt.Errorf("git config safe.directory error: %v", err)
	}

	if err := runGitCommand("config", "--global", "user.name", name); err != nil {
		return fmt.Errorf("git config user.name error: %v", err)
	}

	if err := runGitCommand("config", "--global", "user.email", email); err != nil {
		return fmt.Errorf("git config user.email error: %v", err)
	}

	remoteURL := fmt.Sprintf("https://%s@github.com/%s/%s.git", token, owner, owner)
	if err := runGitCommand("remote", "set-url", "origin", remoteURL); err != nil {
		return fmt.Errorf("git remote set-url error: %v", err)
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
	if branch == "" {
		branch = "main"
	}

	if msg == "" {
		msg = "📝 Update README.md"
	}

	if err := runGitCommand("add", "README.md"); err != nil {
		return err
	}

	if err := runGitCommand("commit", "-m", msg); err != nil {
		return err
	}

	return runGitCommand("push", "origin", branch)
}

func runGitCommand(args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run 'git %v': %v", args, err)
	}

	return nil
}
