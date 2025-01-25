package config

import (
	"log/slog"

	"github.com/spf13/viper"
)

func InitConfig() error {
	viper.SetEnvPrefix("APPENDER")
	viper.AutomaticEnv()
	viper.RegisterAlias("l", "logging")
	return nil
}

func GetLogLevel() (slog.Level, bool) {
	if !viper.IsSet("logging") {
		return slog.LevelInfo, false
	}

	level := viper.GetInt("logging")
	switch level {
	case 1:
		return slog.LevelDebug, true
	case 2:
		return slog.LevelInfo, true
	case 3:
		return slog.LevelWarn, true
	case 4:
		return slog.LevelError, true
	default:
		return slog.LevelInfo, false
	}
}
