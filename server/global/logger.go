package global

import (
	"log/slog"
	"os"
)

var LOG *slog.Logger

func SetupLogger() {
	LOG = slog.New(slog.NewJSONHandler(
		os.Stderr,
		&slog.HandlerOptions{AddSource: true, Level: slog.LevelInfo},
	))
}
