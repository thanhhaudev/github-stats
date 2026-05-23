package writer

import (
	"strings"

	"github.com/thanhhaudev/github-stats/pkg/config"
)

// stripMode is the active STRIP_EMOJI mode for this process. Empty string means
// no stripping (equivalent to config.StripEmojiNone). It is set exactly once
// at process startup via SetStripEmoji and read by the three apply helpers.
var stripMode = ""

// SetStripEmoji sets the active strip mode for subsequent writer calls.
// An empty input is normalized to config.StripEmojiNone. Callers are expected
// to pass a value that has already been validated by config.Validate().
func SetStripEmoji(mode string) {
	if mode == "" {
		stripMode = config.StripEmojiNone
		return
	}
	stripMode = mode
}

// applyTitleStrip removes emoji from s when the active mode targets titles.
func applyTitleStrip(s string) string {
	if stripMode == config.StripEmojiTitle || stripMode == config.StripEmojiAll {
		return stripEmoji(s)
	}
	return s
}

// applyContentStrip removes emoji from s when the active mode targets content
// (text inside ```text code fences).
func applyContentStrip(s string) string {
	if stripMode == config.StripEmojiContent || stripMode == config.StripEmojiAll {
		return stripEmoji(s)
	}
	return s
}

// applyAllStrip removes emoji from s only when the active mode is ALL.
// Used for text that is neither a markdown bold title nor inside a code fence
// (currently just the "Last updated on …" line).
func applyAllStrip(s string) string {
	if stripMode == config.StripEmojiAll {
		return stripEmoji(s)
	}
	return s
}

// stripEmoji removes every rune in the configured emoji ranges from s, then
// collapses every run of two or more ASCII spaces to a single space and trims
// leading/trailing ASCII spaces from each line. Tabs and newlines are left
// intact so multi-line strings keep their structure.
func stripEmoji(s string) string {
	if s == "" {
		return s
	}

	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if isEmojiRune(r) {
			continue
		}
		b.WriteRune(r)
	}

	// Collapse whitespace per-line so a single \n separator survives.
	lines := strings.Split(b.String(), "\n")
	for i, line := range lines {
		// Collapse runs of two or more ASCII spaces to one.
		for strings.Contains(line, "  ") {
			line = strings.ReplaceAll(line, "  ", " ")
		}
		// Trim ASCII space at edges of this line only.
		line = strings.Trim(line, " ")
		lines[i] = line
	}
	return strings.Join(lines, "\n")
}

// isEmojiRune reports whether r is in one of the Unicode ranges that the
// writer emits as emoji or as a modifier of an emoji sequence.
//
// The set is intentionally hand-rolled (no external dependency) and covers
// every emoji literal currently emitted by pkg/writer. It is a strict superset
// of the symbol ranges in isWideRune; any emoji added to a writer literal in
// the future must fall inside one of these ranges or be added here.
func isEmojiRune(r rune) bool {
	switch {
	case r >= 0x1F300 && r <= 0x1FAFF: // Misc Symbols & Pictographs through Symbols & Pictographs Extended-A
		return true
	case r >= 0x2300 && r <= 0x23FF: // Miscellaneous Technical (⏱, ⏳, ⌛, …)
		return true
	case r >= 0x2600 && r <= 0x26FF: // Miscellaneous Symbols (⭐, ☀, …)
		return true
	case r >= 0x2700 && r <= 0x27BF: // Dingbats (✅, ✨, …)
		return true
	case r >= 0x2B00 && r <= 0x2BFF: // Miscellaneous Symbols and Arrows (⬜, ⬛, …)
		return true
	case r >= 0xFE00 && r <= 0xFE0F: // Variation selectors (VS-16 etc.)
		return true
	case r == 0x200D: // Zero-width joiner
		return true
	}
	return false
}
