package main

import (
	"fmt"
	"io"
	"log/slog"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jongschneider/ai-toolbox/tools/appender/config"
)

func setupLogging() error {
	level, shouldLog := config.GetLogLevel()
	if !shouldLog {
		slog.SetDefault(slog.New(slog.NewJSONHandler(io.Discard, nil)))
		return nil
	}

	if err := os.MkdirAll("logs", 0o755); err != nil {
		return fmt.Errorf("could not create logs directory: %w", err)
	}

	logFile, err := tea.LogToFile("logs/debug.log", "debug")
	if err != nil {
		return fmt.Errorf("could not open log file: %w", err)
	}

	logger := slog.New(slog.NewJSONHandler(logFile, &slog.HandlerOptions{
		Level: level,
	}))
	slog.SetDefault(logger)

	return nil
}
