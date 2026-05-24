package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// runGit is a tiny helper for shelling out to git during smoke tests.
func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v in %s: %v\n%s", args, dir, err, out)
	}
}

// TestSetupGitConfig_NoTokenInRemoteURL locks in the security guarantee:
// setupGitConfig must set `origin` to a token-free URL so a leaked
// .git/config cannot expose GITHUB_TOKEN. Auth is supplied per-push via
// http.extraheader injected through GIT_CONFIG_* env vars.
func TestSetupGitConfig_NoTokenInRemoteURL(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available on PATH")
	}

	tmp := t.TempDir()
	work := filepath.Join(tmp, "work")
	if err := os.Mkdir(work, 0o755); err != nil {
		t.Fatal(err)
	}

	runGit(t, work, "init", "-b", "main")
	runGit(t, work, "remote", "add", "origin", "https://example.com/placeholder.git")

	// Sandbox `git config --global` writes so we never touch the user's real
	// ~/.gitconfig.
	t.Setenv("HOME", tmp)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmp, "xdg"))
	t.Chdir(work)

	if err := setupGitConfig("smokeuser", "smoke", "smoke@example.com", false); err != nil {
		t.Fatalf("setupGitConfig: %v", err)
	}

	cfg, err := os.ReadFile(filepath.Join(work, ".git", "config"))
	if err != nil {
		t.Fatal(err)
	}
	s := string(cfg)

	if !strings.Contains(s, "https://github.com/smokeuser/smokeuser.git") {
		t.Errorf("expected clean github URL in .git/config, got:\n%s", s)
	}
	if strings.Contains(s, "@github.com") {
		t.Errorf(".git/config still embeds credentials in the URL:\n%s", s)
	}
}

// TestCommitAndPushReadme_NoTokenLeak exercises the full commit+push flow
// against a local bare repo with a fake token. It asserts:
//  1. the push succeeds (the http.extraheader env vars don't break push)
//  2. the fake token never lands in .git/config
//  3. no `extraheader` config key persists across the push
//
// We deliberately do NOT call setupGitConfig here — it would overwrite
// `origin` with a hard-coded github.com URL.
func TestCommitAndPushReadme_NoTokenLeak(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available on PATH")
	}

	tmp := t.TempDir()
	bare := filepath.Join(tmp, "remote.git")
	work := filepath.Join(tmp, "work")
	if err := os.Mkdir(work, 0o755); err != nil {
		t.Fatal(err)
	}

	runGit(t, tmp, "init", "--bare", "-b", "main", bare)
	runGit(t, work, "init", "-b", "main")
	runGit(t, work, "config", "user.name", "smoke")
	runGit(t, work, "config", "user.email", "smoke@example.com")
	runGit(t, work, "remote", "add", "origin", bare)

	if err := os.WriteFile(filepath.Join(work, "README.md"), []byte("hello\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Chdir(work)

	const fakeToken = "ghp_smoketest_AAAAAAAAAAAAAAAAAAAAAAAAAAA"

	if err := commitAndPushReadme(fakeToken, "smoke test", "main", true); err != nil {
		t.Fatalf("commitAndPushReadme: %v", err)
	}

	cfg, err := os.ReadFile(filepath.Join(work, ".git", "config"))
	if err != nil {
		t.Fatal(err)
	}
	s := string(cfg)

	if strings.Contains(s, fakeToken) {
		t.Errorf(".git/config contains token after push:\n%s", s)
	}
	if strings.Contains(strings.ToLower(s), "extraheader") {
		t.Errorf(".git/config persists extraheader after push:\n%s", s)
	}

	out, err := exec.Command("git", "--git-dir="+bare, "log", "--oneline", "main").Output()
	if err != nil {
		t.Fatalf("bare git log: %v", err)
	}
	if !strings.Contains(string(out), "smoke test") {
		t.Errorf("expected 'smoke test' commit in bare repo, got:\n%s", out)
	}
}
