package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/quiqr/chatcms/internal/auth"
	"github.com/quiqr/chatcms/internal/gitsync"
	"github.com/quiqr/chatcms/internal/httpapi"
	"github.com/quiqr/chatcms/internal/jobs"
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

	queue  *jobs.Queue
	worker *jobs.Worker
	sites  *gitsync.Sites
	auth   *auth.Service
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

	server.sites = gitsync.NewSites(gitsync.Config{
		GitBin:     cfg.GitBin,
		SSHKeyPath: cfg.SSHKeyPath,
		System:     gitsync.Identity{Name: "Quiqr ChatCMS", Email: "chatcms@" + hostOf(cfg.BaseURL)},
	}, cfg.SitesDir())

	server.auth = auth.NewService(db, logger, cfg.BaseURL)

	server.queue = jobs.NewQueue(db)
	server.worker = jobs.NewWorker(server.queue, jobs.WorkerOptions{Logger: logger})
	registerSyncJob(server.worker, server.sites, logger)

	handler := httpapi.New(httpapi.Options{
		Logger: logger,
		Health: server.health,
		Assets: webui.Handler(),
		Auth:   server.auth,
		Secure: strings.HasPrefix(strings.ToLower(cfg.BaseURL), "https://"),
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
	background, stopBackground := context.WithCancel(context.Background())
	defer stopBackground()

	var workers sync.WaitGroup
	workers.Add(2)
	go func() {
		defer workers.Done()
		s.worker.Run(background)
	}()
	go func() {
		defer workers.Done()
		scheduleSyncs(background, s.queue, s.sites, s.cfg.SyncInterval, s.logger)
	}()

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
		stopBackground()
		workers.Wait()
		s.db.Close()
		return err
	case <-ctx.Done():
	}

	s.logger.Info("shutting down", "timeout", s.cfg.ShutdownTimeout)

	shutdownErr := gracefulShutdown(s.http, s.cfg.ShutdownTimeout)
	<-errs

	stopBackground()
	if err := s.worker.Drain(context.Background(), s.cfg.ShutdownTimeout); err != nil {
		s.logger.Error("returning unfinished work to the queue failed", "error", err)
	}
	workers.Wait()

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

// Auth exposes the identity service, which the integration tests use to
// create a person without an endpoint for it.
func (s *Server) Auth() *auth.Service { return s.auth }

// hostOf is used for the address versions are written by, so history shows
// where an edit came from rather than a placeholder domain.
func hostOf(baseURL string) string {
	parsed, err := url.Parse(baseURL)
	if err != nil || parsed.Hostname() == "" {
		return "chatcms.local"
	}
	return strings.ToLower(parsed.Hostname())
}
