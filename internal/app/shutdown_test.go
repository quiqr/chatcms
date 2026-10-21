package app

import (
	"errors"
	"io"
	"net"
	"net/http"
	"testing"
	"time"
)

func TestGracefulShutdownWaitsForAnIdleServer(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	server := &http.Server{Handler: http.NotFoundHandler()}
	go server.Serve(listener)

	if err := gracefulShutdown(server, time.Second); err != nil {
		t.Errorf("gracefulShutdown() error = %v, want nil", err)
	}
}

func TestGracefulShutdownIsForcedWhenARequestOutlivesTheTimeout(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	entered := make(chan struct{})
	release := make(chan struct{})
	defer close(release)

	server := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(entered)
		<-release
	})}
	go server.Serve(listener)

	go func() {
		response, err := http.Get("http://" + listener.Addr().String() + "/slow")
		if err == nil {
			io.Copy(io.Discard, response.Body)
			response.Body.Close()
		}
	}()

	<-entered

	err = gracefulShutdown(server, 50*time.Millisecond)
	if !errors.Is(err, ErrShutdownForced) {
		t.Errorf("gracefulShutdown() error = %v, want ErrShutdownForced", err)
	}
}
