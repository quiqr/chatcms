package app

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
)

func TestSchedulingStopsWithoutAnInterval(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	done := make(chan struct{})
	go func() {
		scheduleSyncs(context.Background(), nil, nil, 0, logger)
		close(done)
	}()

	<-done

	if !strings.Contains(buf.String(), "sync interval") {
		t.Errorf("log = %q, want it to say why nothing will be scheduled", buf.String())
	}
}
