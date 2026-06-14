package language

import (
	"testing"

	"github.com/thanhhaudev/github-stats/pkg/github"
)

func TestFromFilesAggregatesAdditionsAndSkipsVendorGenerated(t *testing.T) {
	files := []github.CommitFile{
		{Filename: "cmd/main.go", Additions: 10},
		{Filename: "web/style.css", Additions: 4},
		{Filename: "node_modules/left-pad/index.js", Additions: 100}, // vendored
		{Filename: "web/app.min.js", Additions: 50},                  // generated (minified)
		{Filename: "cmd/extra.go", Additions: 5},
		{Filename: "data.bin", Additions: 7},    // unknown extension
		{Filename: "config.json", Additions: 9}, // data format, intentionally not counted
	}

	got := FromFiles(files)

	if got["Go"] != 15 {
		t.Errorf("Go additions = %d, want 15", got["Go"])
	}
	if got["CSS"] != 4 {
		t.Errorf("CSS additions = %d, want 4", got["CSS"])
	}
	if _, ok := got["JavaScript"]; ok {
		t.Errorf("expected vendored/minified JS to be excluded, got %d", got["JavaScript"])
	}
	if len(got) != 2 {
		t.Errorf("expected exactly Go and CSS, got %v", got)
	}
}

func TestColorReturnsHexAndFallback(t *testing.T) {
	if c := Color("Go"); c != "#00ADD8" {
		t.Errorf("Color(Go) = %q, want #00ADD8", c)
	}
	if c := Color("NotARealLanguage___"); c != "#ededed" {
		t.Errorf("Color fallback = %q, want #ededed", c)
	}
}
