package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/quiqr/chatcms/internal/httpapi"
	"github.com/quiqr/chatcms/internal/store"
	"github.com/quiqr/chatcms/internal/version"
	"github.com/quiqr/chatcms/internal/webui"
)

var ErrShutdownForced = errors.New("in-flight requests did not finish within the shutdown timeout")

type Server struct {
	cfg      Config
	logger   *slog.Logger
	db       *store.DB
	listener net.Listener
	http     *http.Server
	known    int64
}

func NewServer(ctx context.Context, cfg Config, logger *slog.Logger) (*Server, error) {
	db, err := store.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("the database is not reachable: %w", err)
	}

	if err := db.Migrate(ctx); err != nil {
		db.Close()
		return nil, err
	}

	known, err := db.KnownSchemaVersion()
	if err != nil {
		db.Close()
		return nil, err
	}

	server := &Server{cfg: cfg, logger: logger, db: db, known: known}

	handler := httpapi.New(httpapi.Options{
		Logger: logger,
		Health: server.health,
		Assets: webui.Handler(),
	})

	listener, err := net.Listen("tcp", cfg.Listen)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("listening on %s failed: %w", cfg.Listen, err)
	}

	server.listener = listener
	server.http = &http.Server{
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
	}
	return server, nil
}

func (s *Server) Addr() string {
	return s.listener.Addr().String()
}

func (s *Server) health(ctx context.Context) httpapi.Health {
	health := httpapi.Health{
		Status:        httpapi.HealthStatusOK,
		ServerVersion: version.Version,
		Database:      httpapi.DatabaseReachable,
		SchemaVersion: fmt.Sprintf("%d", s.known),
	}

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if err := s.db.Ping(ctx); err != nil {
		health.Status = httpapi.HealthStatusDegraded
		health.Database = httpapi.DatabaseUnreachable
		return health
	}

	applied, err := s.db.SchemaVersion(ctx)
	if err != nil {
		health.Status = httpapi.HealthStatusDegraded
		health.Database = httpapi.DatabaseUnreachable
		return health
	}
	health.SchemaVersion = fmt.Sprintf("%d", applied)
	return health
}

func (s *Server) Serve(ctx context.Context) error {
	errs := make(chan error, 1)
	go func() {
		err := s.http.Serve(s.listener)
		if errors.Is(err, http.ErrServerClosed) {
			err = nil
		}
		errs <- err
	}()

	s.logger.Info("listening", "address", s.Addr(), "version", version.Version)

	select {
	case err := <-errs:
		s.db.Close()
		return err
	case <-ctx.Done():
	}

	s.logger.Info("shutting down", "timeout", s.cfg.ShutdownTimeout)

	shutdownErr := gracefulShutdown(s.http, s.cfg.ShutdownTimeout)
	<-errs
	s.db.Close()

	if errors.Is(shutdownErr, ErrShutdownForced) {
		s.logger.Error("shutdown was forced", "timeout", s.cfg.ShutdownTimeout)
	}
	return shutdownErr
}

func (s *Server) Close() error {
	if s.db != nil {
		s.db.Close()
	}
	return s.http.Close()
}

func Serve(ctx context.Context, cfg Config, logger *slog.Logger) error {
	server, err := NewServer(ctx, cfg, logger)
	if err != nil {
		return err
	}
	return server.Serve(ctx)
}
