package main

import (
	"fmt"
	"log/slog"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func setupLogging() error {
	// Create logs directory if it doesn't exist
	if err := os.MkdirAll("logs", 0o755); err != nil {
		return fmt.Errorf("could not create logs directory: %w", err)
	}

	// Open log file
	logFile, err := tea.LogToFile("logs/debug.log", "debug")
	if err != nil {
		return fmt.Errorf("could not open log file: %w", err)
	}

	// Configure slog
	logger := slog.New(slog.NewJSONHandler(logFile, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(logger)

	return nil
}
