package writer

import (
	"testing"

	"github.com/thanhhaudev/github-stats/pkg/config"
)

func setStripModeForTest(t *testing.T, mode string) {
	t.Helper()
	prev := stripMode
	stripMode = mode
	t.Cleanup(func() { stripMode = prev })
}

func TestStripEmoji(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"empty", "", ""},
		{"plain text", "plain text", "plain text"},
		{"single codepoint leading", "📊 Foo", "Foo"},
		{"single codepoint trailing", "Foo 🐤", "Foo"},
		{"both ends", "🕒 I'm An Early Bird 🐤", "I'm An Early Bird"},
		{"emoji + VS-16", "⏱️ All Time Stats", "All Time Stats"},
		{"ZWJ + skin tone", "🥷🏻 ninja", "ninja"},
		{"middle emoji collapses spaces", "a 📊 b", "a b"},
		{"emoji only", "📊", ""},
		{"multiple emoji middle", "a 📊 🐤 b", "a b"},
		{"dingbat (✅)", "✅ ok", "ok"},
		{"misc symbol (⏳)", "⏳ later", "later"},
		{"keeps newline separator", "line1\n📊 line2", "line1\nline2"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stripEmoji(tt.in)
			if got != tt.want {
				t.Errorf("stripEmoji(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestApplyTitleStrip(t *testing.T) {
	cases := []struct {
		mode string
		in   string
		want string
	}{
		{config.StripEmojiNone, "📈 Coding Streak", "📈 Coding Streak"},
		{config.StripEmojiTitle, "📈 Coding Streak", "Coding Streak"},
		{config.StripEmojiContent, "📈 Coding Streak", "📈 Coding Streak"},
		{config.StripEmojiAll, "📈 Coding Streak", "Coding Streak"},
		{"", "📈 Coding Streak", "📈 Coding Streak"},
	}
	for _, c := range cases {
		t.Run(c.mode+"/"+c.in, func(t *testing.T) {
			setStripModeForTest(t, c.mode)
			if got := applyTitleStrip(c.in); got != c.want {
				t.Errorf("applyTitleStrip(%q) under mode %q = %q, want %q", c.in, c.mode, got, c.want)
			}
		})
	}
}

func TestApplyContentStrip(t *testing.T) {
	cases := []struct {
		mode string
		in   string
		want string
	}{
		{config.StripEmojiNone, "🔥 Current Streak:", "🔥 Current Streak:"},
		{config.StripEmojiTitle, "🔥 Current Streak:", "🔥 Current Streak:"},
		{config.StripEmojiContent, "🔥 Current Streak:", "Current Streak:"},
		{config.StripEmojiAll, "🔥 Current Streak:", "Current Streak:"},
	}
	for _, c := range cases {
		t.Run(c.mode+"/"+c.in, func(t *testing.T) {
			setStripModeForTest(t, c.mode)
			if got := applyContentStrip(c.in); got != c.want {
				t.Errorf("applyContentStrip(%q) under mode %q = %q, want %q", c.in, c.mode, got, c.want)
			}
		})
	}
}

func TestApplyAllStrip(t *testing.T) {
	cases := []struct {
		mode string
		in   string
		want string
	}{
		{config.StripEmojiNone, "⏳ *Last updated on 2026-01-01*", "⏳ *Last updated on 2026-01-01*"},
		{config.StripEmojiTitle, "⏳ *Last updated on 2026-01-01*", "⏳ *Last updated on 2026-01-01*"},
		{config.StripEmojiContent, "⏳ *Last updated on 2026-01-01*", "⏳ *Last updated on 2026-01-01*"},
		{config.StripEmojiAll, "⏳ *Last updated on 2026-01-01*", "*Last updated on 2026-01-01*"},
	}
	for _, c := range cases {
		t.Run(c.mode+"/"+c.in, func(t *testing.T) {
			setStripModeForTest(t, c.mode)
			if got := applyAllStrip(c.in); got != c.want {
				t.Errorf("applyAllStrip(%q) under mode %q = %q, want %q", c.in, c.mode, got, c.want)
			}
		})
	}
}

func TestSetStripEmoji_NormalizesEmpty(t *testing.T) {
	setStripModeForTest(t, "leftover")
	SetStripEmoji("")
	if stripMode != config.StripEmojiNone {
		t.Errorf("expected SetStripEmoji(\"\") to normalize to %q, got %q", config.StripEmojiNone, stripMode)
	}
}
