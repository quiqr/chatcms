package app

import (
	"context"
	"errors"
	"net/http"
	"time"
)

func gracefulShutdown(server *http.Server, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	err := server.Shutdown(ctx)
	if errors.Is(err, context.DeadlineExceeded) {
		return ErrShutdownForced
	}
	return err
}
