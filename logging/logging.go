// Package logging configures the game's structured logger: log/slog writing to
// a dated, rotating file under logs/. The on-screen CLI battle output stays on
// stdout (fmt); this is the persistent record for debugging and grading.
package logging

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
)

const logDir = "logs"

// parseLevel maps a LOG_LEVEL string to an slog.Level. Empty or unknown values
// fall back to Info.
func parseLevel(s string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// logFilePath returns the dated log-file path for the given time,
// e.g. logs/battle-2026-07-05.log.
func logFilePath(now time.Time) string {
	return filepath.Join(logDir, "battle-"+now.Format("2006-01-02")+".log")
}

// Init configures the global slog logger to write JSON logs to a dated,
// size-rotating file under logs/. The level comes from LOG_LEVEL
// (debug|info|warn|error, default info). Call once at startup.
func Init() error {
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return fmt.Errorf("create log dir: %w", err)
	}
	writer := &lumberjack.Logger{
		Filename:   logFilePath(time.Now()),
		MaxSize:    5, // MB before a rotation
		MaxBackups: 5,
		MaxAge:     30, // days
		Compress:   true,
	}
	handler := slog.NewJSONHandler(writer, &slog.HandlerOptions{
		Level: parseLevel(os.Getenv("LOG_LEVEL")),
	})
	slog.SetDefault(slog.New(handler))
	return nil
}
