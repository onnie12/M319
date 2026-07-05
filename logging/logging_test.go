package logging

import (
	"log/slog"
	"path/filepath"
	"testing"
	"time"
)

func TestParseLevel(t *testing.T) {
	cases := []struct {
		in   string
		want slog.Level
	}{
		{"debug", slog.LevelDebug},
		{"DEBUG", slog.LevelDebug},
		{"info", slog.LevelInfo},
		{"warn", slog.LevelWarn},
		{"warning", slog.LevelWarn},
		{"error", slog.LevelError},
		{"", slog.LevelInfo},         // default
		{"nonsense", slog.LevelInfo}, // unknown -> default
	}
	for _, c := range cases {
		if got := parseLevel(c.in); got != c.want {
			t.Errorf("parseLevel(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestLogFilePath_UsesDatedName(t *testing.T) {
	now := time.Date(2026, 7, 5, 12, 0, 0, 0, time.UTC)
	got := logFilePath(now)
	want := filepath.Join("logs", "battle-2026-07-05.log")
	if got != want {
		t.Errorf("logFilePath = %q, want %q", got, want)
	}
}
