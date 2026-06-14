// Package language maps changed files to programming languages using a small,
// hand-maintained extension table (no external dependency). It mirrors the
// spirit of GitHub Linguist's defaults: vendored and generated files do not
// count toward a user's stats, and pure-data formats (JSON/YAML/...) are not
// treated as "languages I code in". Add a language by editing the table below.
package language

import (
	"path/filepath"
	"strings"

	"github.com/thanhhaudev/github-stats/pkg/github"
)

type langDef struct {
	Name  string
	Color string // Linguist hex color, including leading '#'
	Exts  []string
}

// languages is the single source of truth. Both the extension index and the
// name->color index are derived from it at init time.
var languages = []langDef{
	{"Go", "#00ADD8", []string{".go"}},
	{"JavaScript", "#f1e05a", []string{".js", ".mjs", ".cjs", ".jsx"}},
	{"TypeScript", "#3178c6", []string{".ts", ".tsx", ".mts", ".cts"}},
	{"Python", "#3572A5", []string{".py", ".pyw"}},
	{"Java", "#b07219", []string{".java"}},
	{"Kotlin", "#A97BFF", []string{".kt", ".kts"}},
	// .h (C/C++/Obj-C) and .m (Obj-C/MATLAB) are ambiguous without content
	// analysis, so they are intentionally omitted: under-counting an obscure
	// header is safer than crediting a language the user never wrote.
	{"C", "#555555", []string{".c"}},
	{"C++", "#f34b7d", []string{".cpp", ".cc", ".cxx", ".c++", ".hpp", ".hh", ".hxx"}},
	{"C#", "#178600", []string{".cs"}},
	{"Ruby", "#701516", []string{".rb"}},
	{"PHP", "#4F5D95", []string{".php"}},
	{"Swift", "#F05138", []string{".swift"}},
	{"Objective-C++", "#6866fb", []string{".mm"}},
	{"Scala", "#c22d40", []string{".scala", ".sc"}},
	{"Rust", "#dea584", []string{".rs"}},
	{"Dart", "#00B4AB", []string{".dart"}},
	{"Elixir", "#6e4a7e", []string{".ex", ".exs"}},
	{"Erlang", "#B83998", []string{".erl"}},
	{"Haskell", "#5e5086", []string{".hs"}},
	{"Lua", "#000080", []string{".lua"}},
	{"Perl", "#0298c3", []string{".pl", ".pm"}},
	{"R", "#198CE7", []string{".r"}},
	{"Julia", "#a270ba", []string{".jl"}},
	{"Clojure", "#db5855", []string{".clj", ".cljs", ".cljc", ".edn"}},
	{"Groovy", "#4298b8", []string{".groovy", ".gradle"}},
	{"Shell", "#89e051", []string{".sh", ".bash", ".zsh"}},
	{"PowerShell", "#012456", []string{".ps1", ".psm1"}},
	{"Vim Script", "#199f4b", []string{".vim"}},
	{"SQL", "#e38c00", []string{".sql"}},
	{"HTML", "#e34c26", []string{".html", ".htm"}},
	{"CSS", "#563d7c", []string{".css"}},
	{"SCSS", "#c6538c", []string{".scss"}},
	{"Sass", "#a53b70", []string{".sass"}},
	{"Less", "#1d365d", []string{".less"}},
	{"Vue", "#41b883", []string{".vue"}},
	{"Svelte", "#ff3e00", []string{".svelte"}},
	{"Astro", "#ff5a03", []string{".astro"}},
	{"Markdown", "#083fa1", []string{".md", ".markdown"}},
	{"Solidity", "#AA6746", []string{".sol"}},
	{"Zig", "#ec915c", []string{".zig"}},
	{"Nim", "#ffc200", []string{".nim"}},
	{"OCaml", "#3be133", []string{".ml", ".mli"}},
	{"F#", "#b845fc", []string{".fs", ".fsx"}},
	{"Assembly", "#6E4C13", []string{".asm", ".s"}},
}

var (
	byExt       = map[string]langDef{}
	colorByName = map[string]string{}
)

func init() {
	for _, l := range languages {
		colorByName[l.Name] = l.Color
		for _, e := range l.Exts {
			byExt[e] = l
		}
	}
}

// vendorSegments are path components that mark third-party code the user did not
// author. Matched as whole path segments to avoid false positives.
var vendorSegments = map[string]struct{}{
	"node_modules":     {},
	"vendor":           {},
	"bower_components": {},
	"dist":             {},
	"third_party":      {},
	".yarn":            {},
	"pods":             {}, // CocoaPods
	"carthage":         {}, // Carthage
	"external":         {}, // Bazel external deps
	"deps":             {}, // Elixir / general dependency dirs
}

func isVendor(path string) bool {
	for _, seg := range strings.Split(strings.ToLower(path), "/") {
		if _, ok := vendorSegments[seg]; ok {
			return true
		}
	}

	return false
}

func isGenerated(path string) bool {
	p := strings.ToLower(path)
	switch {
	case strings.HasSuffix(p, ".min.js"), strings.HasSuffix(p, ".min.css"):
		return true
	case strings.HasSuffix(p, ".pb.go"), strings.HasSuffix(p, "_pb2.py"):
		return true
	case strings.HasSuffix(p, ".designer.cs"): // .NET designer-generated
		return true
	case strings.Contains(p, "_generated."):
		return true
	}

	for _, seg := range strings.Split(p, "/") {
		if seg == "generated" || seg == "generated-sources" {
			return true
		}
	}

	return false
}

// FromFiles returns language -> additions for the given commit files. Vendored
// and generated files are skipped; files whose extension is not in the table
// (binaries, data, config) are ignored.
func FromFiles(files []github.CommitFile) map[string]int {
	out := make(map[string]int)
	for _, f := range files {
		if isVendor(f.Filename) || isGenerated(f.Filename) {
			continue
		}

		ext := strings.ToLower(filepath.Ext(f.Filename))
		def, ok := byExt[ext]
		if !ok {
			continue
		}

		out[def.Name] += f.Additions
	}

	return out
}

// Color returns the hex color for a language, with a neutral fallback so writers
// can always strip the leading '#'.
func Color(name string) string {
	if c, ok := colorByName[name]; ok {
		return c
	}

	return "#ededed"
}
