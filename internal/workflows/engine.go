package workflows

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
	DefaultCheckTimeout = 5 * time.Minute
	DefaultCheckMemory  = 1 << 30
	MaxDetailBytes      = 64 << 10
)

type SiteBuilder interface {
	Build(ctx context.Context, site fs.FS, options sitebuild.Options) (*runner.Result, error)
}

// Publisher sends a project's work to its remote. It is given what a build
// produced when one ran, and nothing when the project publishes its source.
// What lands where is the project's publishing settings, not this package's.
//
// What it is given is only readable while the call lasts: the workspace a
// build used is released when the workflow ends. A publisher that wants to
// keep the pages has to copy them.
type Publisher interface {
	Publish(ctx context.Context, projectID uuid.UUID, built fs.FS) error
}

type Engine struct {
	runner    runner.Runner
	builder   SiteBuilder
	publisher Publisher
	toolbox   []string
	limits    Limits
}

type Limits struct {
	CheckTimeout time.Duration
	CheckMemory  int64
}

func (l Limits) withDefaults() Limits {
	if l.CheckTimeout <= 0 {
		l.CheckTimeout = DefaultCheckTimeout
	}
	if l.CheckMemory <= 0 {
		l.CheckMemory = DefaultCheckMemory
	}
	return l
}

var ErrIncomplete = errors.New("workflows: the engine is missing a part it needs")

func NewEngine(r runner.Runner, builder SiteBuilder, publisher Publisher, toolbox []string, limits Limits) (*Engine, error) {
	if r == nil || builder == nil || publisher == nil {
		return nil, ErrIncomplete
	}
	return &Engine{
		runner:    r,
		builder:   builder,
		publisher: publisher,
		toolbox:   append([]string{}, toolbox...),
		limits:    limits.withDefaults(),
	}, nil
}

type Run struct {
	ProjectID uuid.UUID
	Site      fs.FS
	BaseURL   string
}

func (e *Engine) Run(ctx context.Context, workflow Workflow, run Run) (*Outcome, error) {
	if run.Site == nil {
		return nil, errors.New("workflows: there is no site to run against")
	}

	outcome := &Outcome{Trigger: workflow.Trigger}

	var built *runner.Result
	defer func() { built.Release() }()

	for _, step := range workflow.Steps {
		started := time.Now()
		done := StepOutcome{Title: step.Title(), Action: step.Action, State: Succeeded}

		var trouble *Trouble
		var err error

		switch step.Action {
		case Check:
			done.Detail, trouble, err = e.check(ctx, step, run)
		case Build:
			built, done.Detail, trouble, err = e.build(ctx, run)
		case Publish:
			trouble, err = e.publish(ctx, run, built)
		default:
			err = fmt.Errorf("workflows: %q is not something this engine can do", step.Action)
		}

		done.Duration = time.Since(started)
		done.Detail = runner.TrimLog(done.Detail, MaxDetailBytes)
		if err != nil {
			return nil, err
		}
		if trouble != nil {
			trouble.Step = done.Title
			done.State = Failed
			outcome.Steps = append(outcome.Steps, done)
			outcome.Trouble = trouble
			return outcome, nil
		}
		outcome.Steps = append(outcome.Steps, done)
	}
	return outcome, nil
}

func (e *Engine) check(ctx context.Context, step Step, run Run) (string, *Trouble, error) {
	timeout := step.Timeout
	if timeout <= 0 {
		timeout = e.limits.CheckTimeout
	}

	result, err := e.runner.Run(ctx, runner.Task{
		Inputs:      run.Site,
		Command:     "sh",
		Arguments:   []string{"-c", step.Run},
		Tools:       e.toolbox,
		Timeout:     timeout,
		MemoryBytes: e.limits.CheckMemory,
	})
	if err != nil {
		if errors.Is(err, runner.ErrNoIsolation) {
			return "", &Trouble{
				Because: "This site's own check could not be run safely, so we stopped.",
				Advice:  "Your implementer has been told about this.",
			}, nil
		}
		return "", nil, err
	}
	defer result.Release()

	if result.Succeeded() {
		return result.Log, nil, nil
	}
	return result.Log, checkTrouble(result), nil
}

func checkTrouble(result *runner.Result) *Trouble {
	switch result.Reached {
	case runner.LimitTime:
		return &Trouble{
			Because: "This site's own check took too long and was stopped.",
			Advice:  "Your implementer has been told about this.",
		}
	case runner.LimitMemory:
		return &Trouble{
			Because: "This site's own check needed more memory than it is allowed.",
			Advice:  "Your implementer has been told about this.",
		}
	default:
		return &Trouble{
			Because: "A check on this site found something that needs fixing first.",
			Advice:  "Your implementer has been told about this.",
		}
	}
}

func (e *Engine) build(ctx context.Context, run Run) (*runner.Result, string, *Trouble, error) {
	result, err := e.builder.Build(ctx, run.Site, sitebuild.Options{BaseURL: run.BaseURL})
	if err != nil {
		return nil, "", &Trouble{
			Because: "The site could not be built.",
			Advice:  "Your implementer has been told about this.",
		}, nil
	}
	if !result.Succeeded() {
		detail := result.Log
		result.Release()
		return nil, detail, &Trouble{
			Because: "The site could not be built.",
			Advice:  "Your implementer has been told about this.",
		}, nil
	}
	return result, result.Log, nil, nil
}

func (e *Engine) publish(ctx context.Context, run Run, built *runner.Result) (*Trouble, error) {
	var pages fs.FS
	if built != nil {
		pages = sitebuild.Pages(built)
	}

	if err := e.publisher.Publish(ctx, run.ProjectID, pages); err != nil {
		return &Trouble{
			Because: "We could not reach the place your site is published to.",
			Advice:  "Try again in a moment. If it keeps happening, your implementer can look into it.",
		}, nil
	}
	return nil, nil
}
