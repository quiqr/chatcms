package app

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/quiqr/chatcms/internal/version"
)

func Run(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("chatcmsd", flag.ContinueOnError)
	fs.SetOutput(stderr)
	showVersion := fs.Bool("version", false, "print the version and exit")

	if err := fs.Parse(args); err != nil {
		return 2
	}

	if *showVersion {
		fmt.Fprintln(stdout, version.String())
		return 0
	}

	cfg, err := LoadConfig(EnvLookup())
	if err != nil {
		fmt.Fprintf(stderr, "chatcmsd: %v\n", err)
		return 1
	}

	logger := NewLogger(stderr, cfg.LogFormat, IsTerminal(stderr))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := Serve(ctx, cfg, logger); err != nil {
		if errors.Is(err, ErrShutdownForced) {
			logger.Error("shutdown was forced", "error", err)
			return 1
		}
		logger.Error("the server stopped", "error", err)
		return 1
	}
	return 0
}

func NewLogger(w io.Writer, format string, terminal bool) *slog.Logger {
	opts := &slog.HandlerOptions{Level: slog.LevelInfo}

	useText := format == "text" || (format == "auto" && terminal)
	if useText {
		return slog.New(slog.NewTextHandler(w, opts))
	}
	return slog.New(slog.NewJSONHandler(w, opts))
}

func IsTerminal(w io.Writer) bool {
	file, ok := w.(*os.File)
	if !ok {
		return false
	}
	info, err := file.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}
