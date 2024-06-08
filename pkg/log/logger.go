package log

import (
	"log/slog"
	"os"

	"github.com/charmbracelet/log"
)

type Options struct {
	Level      slog.Level
	ShowTime   bool
	ShowCaller bool
	Prefix     string
}

func NewLogger(opts Options) *slog.Logger {
	logger := slog.New(log.NewWithOptions(os.Stderr, log.Options{
		Level:           log.Level(opts.Level),
		Prefix:          opts.Prefix,
		ReportTimestamp: opts.ShowTime,
		ReportCaller:    opts.ShowCaller,
	}))
	return logger
}
