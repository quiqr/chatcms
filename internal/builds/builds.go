package builds

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"time"

	"github.com/google/uuid"

	"github.com/quiqr/chatcms/internal/runner"
	"github.com/quiqr/chatcms/internal/sitebuild"
)

const (
	MaxLogBytes       = 256 << 10
	DefaultListLength = 50
	interruptedNote   = "This build was still running when the server stopped, so what happened to it is not known."
)

type State string

const (
	Running     State = "running"
	Succeeded   State = "succeeded"
	Failed      State = "failed"
	OutOfTime   State = "out-of-time"
	OutOfMemory State = "out-of-memory"
	Interrupted State = "interrupted"
)

func (s State) Finished() bool { return s != Running }

type Build struct {
	ID         uuid.UUID
	ProjectID  uuid.UUID
	Reason     string
	State      State
	ExitCode   *int32
	Log        string
	StartedAt  time.Time
	FinishedAt *time.Time
}

func (b Build) Duration() time.Duration {
	if b.FinishedAt == nil {
		return 0
	}
	return b.FinishedAt.Sub(b.StartedAt)
}

type Recorder interface {
	Start(ctx context.Context, projectID uuid.UUID, reason string) (Build, error)
	Finish(ctx context.Context, id uuid.UUID, state State, exitCode *int32, log string) (Build, error)
	Get(ctx context.Context, id uuid.UUID) (Build, error)
	List(ctx context.Context, projectID uuid.UUID, limit int32) ([]Build, error)
	InterruptLeftovers(ctx context.Context) (int, error)
}

type Siteloader func(ctx context.Context, projectID uuid.UUID) (fs.FS, error)

type SiteBuilder interface {
	Build(ctx context.Context, site fs.FS, options sitebuild.Options) (*runner.Result, error)
}

type Outcome struct {
	Build   Build
	Outputs fs.FS

	release func() error
}

func (o *Outcome) Release() error {
	if o == nil || o.release == nil {
		return nil
	}
	release := o.release
	o.release = nil
	o.Outputs = nil
	return release()
}

type Service struct {
	records Recorder
	builder SiteBuilder
	sites   Siteloader
}

var ErrIncomplete = errors.New("builds: the service is missing a part it needs")

func NewService(records Recorder, builder SiteBuilder, sites Siteloader) (*Service, error) {
	if records == nil || builder == nil || sites == nil {
		return nil, ErrIncomplete
	}
	return &Service{records: records, builder: builder, sites: sites}, nil
}

func (s *Service) Run(ctx context.Context, projectID uuid.UUID, reason string, options sitebuild.Options) (*Outcome, error) {
	record, err := s.records.Start(ctx, projectID, reason)
	if err != nil {
		return nil, err
	}

	site, err := s.sites(ctx, projectID)
	if err != nil {
		return s.failed(ctx, record, "The site could not be read: "+err.Error())
	}

	result, err := s.builder.Build(ctx, site, options)
	if err != nil {
		return s.failed(ctx, record, "The build could not be started: "+err.Error())
	}

	exitCode := int32(result.ExitCode)
	finished, err := s.records.Finish(ctx, record.ID, StateOf(result), &exitCode, runner.TrimLog(result.Log, MaxLogBytes))
	if err != nil {
		result.Release()
		return nil, err
	}
	if finished.State != Succeeded {
		result.Release()
		return &Outcome{Build: finished}, nil
	}
	return &Outcome{Build: finished, Outputs: result.Outputs, release: result.Release}, nil
}

func (s *Service) failed(ctx context.Context, record Build, why string) (*Outcome, error) {
	finished, err := s.records.Finish(ctx, record.ID, Failed, nil, runner.TrimLog(why, MaxLogBytes))
	if err != nil {
		return &Outcome{Build: record}, nil
	}
	return &Outcome{Build: finished}, nil
}

func StateOf(result *runner.Result) State {
	switch {
	case result.Reached == runner.LimitTime:
		return OutOfTime
	case result.Reached == runner.LimitMemory:
		return OutOfMemory
	case result.ExitCode != 0:
		return Failed
	default:
		return Succeeded
	}
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (Build, error) {
	return s.records.Get(ctx, id)
}

func (s *Service) List(ctx context.Context, projectID uuid.UUID, limit int32) ([]Build, error) {
	if limit <= 0 {
		limit = DefaultListLength
	}
	return s.records.List(ctx, projectID, limit)
}

func (s *Service) InterruptLeftovers(ctx context.Context) (int, error) {
	count, err := s.records.InterruptLeftovers(ctx)
	if err != nil {
		return 0, fmt.Errorf("builds: marking leftover builds as interrupted: %w", err)
	}
	return count, nil
}

func TrimToLimit(log string) string {
	return runner.TrimLog(log, MaxLogBytes)
}
