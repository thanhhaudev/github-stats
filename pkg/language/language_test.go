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

func TestFromFilesDropsAmbiguousExtensions(t *testing.T) {
	files := []github.CommitFile{
		{Filename: "src/foo.h", Additions: 100}, // ambiguous C/C++/Obj-C -> dropped
		{Filename: "src/bar.m", Additions: 100}, // ambiguous Obj-C/MATLAB -> dropped
		{Filename: "src/baz.c", Additions: 5},   // unambiguous C
		{Filename: "src/qux.mm", Additions: 7},  // unambiguous Objective-C++
	}

	got := FromFiles(files)

	if got["C"] != 5 {
		t.Errorf("C = %d, want 5 (only .c counted, .h dropped)", got["C"])
	}
	if got["Objective-C++"] != 7 {
		t.Errorf("Objective-C++ = %d, want 7", got["Objective-C++"])
	}
	if _, ok := got["Objective-C"]; ok {
		t.Errorf("ambiguous .m must not be credited to Objective-C, got %v", got)
	}
	if len(got) != 2 {
		t.Errorf("expected only C and Objective-C++, got %v", got)
	}
}

func TestFromFilesExcludesMoreVendorAndGenerated(t *testing.T) {
	files := []github.CommitFile{
		{Filename: "Pods/Lib/Source.swift", Additions: 100},             // vendored
		{Filename: "Carthage/Build/x.swift", Additions: 100},            // vendored
		{Filename: "external/dep/a.go", Additions: 100},                 // vendored
		{Filename: "deps/elixir_thing/b.go", Additions: 100},            // vendored
		{Filename: "app/Form.designer.cs", Additions: 100},              // generated
		{Filename: "target/generated-sources/Gen.java", Additions: 100}, // generated
		{Filename: "src/real.go", Additions: 8},                         // counted
	}

	got := FromFiles(files)

	if got["Go"] != 8 {
		t.Errorf("Go = %d, want 8 (only src/real.go)", got["Go"])
	}
	if len(got) != 1 {
		t.Errorf("expected only Go after exclusions, got %v", got)
	}
}
