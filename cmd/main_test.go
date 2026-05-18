package main

import (
	"os"
	"strings"
	"testing"
)

func TestUpdateReadmeReplacesOnlyConfiguredSection(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	original := strings.Join([]string{
		"# Profile",
		"",
		"keep before",
		"<!--START_SECTION:readme-stats-->",
		"old stats",
		"<!--END_SECTION:readme-stats-->",
		"keep after",
		"<!--START_SECTION:other-->",
		"other content",
		"<!--END_SECTION:other-->",
		"",
	}, "\n")
	if err := os.WriteFile("README.md", []byte(original), 0644); err != nil {
		t.Fatal(err)
	}

	if err := updateReadme("new stats", "readme-stats"); err != nil {
		t.Fatalf("updateReadme returned error: %v", err)
	}

	b, err := os.ReadFile("README.md")
	if err != nil {
		t.Fatal(err)
	}
	got := string(b)
	for _, want := range []string{"keep before", "new stats", "keep after", "other content"} {
		if !strings.Contains(got, want) {
			t.Fatalf("updated README missing %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, "old stats") {
		t.Fatalf("updated README still contains old stats:\n%s", got)
	}
}

func TestUpdateReadmeReturnsErrorWhenSectionMissing(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	if err := os.WriteFile("README.md", []byte("# Profile\n"), 0644); err != nil {
		t.Fatal(err)
	}

	err := updateReadme("new stats", "readme-stats")
	if err == nil {
		t.Fatal("expected missing section error, got nil")
	}
	if !strings.Contains(err.Error(), "section tags") {
		t.Fatalf("expected section tag error, got %v", err)
	}
}
