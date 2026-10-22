package workflows

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

type Trigger string

const (
	OnSave    Trigger = "save"
	OnPreview Trigger = "preview"
	OnPublish Trigger = "publish"
)

var Triggers = []Trigger{OnSave, OnPreview, OnPublish}

func (t Trigger) Known() bool {
	for _, known := range Triggers {
		if t == known {
			return true
		}
	}
	return false
}

type Action string

const (
	Build   Action = "build"
	Check   Action = "check"
	Publish Action = "publish"
)

var Actions = []Action{Build, Check, Publish}

func (a Action) Known() bool {
	for _, known := range Actions {
		if a == known {
			return true
		}
	}
	return false
}

// Wording is what a person is shown for a step that was not given a name. The
// words are the ones docs/design/DESIGN.md section 5.8 uses for the publish
// steps, so a default workflow reads as the design describes it.
var wording = map[Action]string{
	Build:   "Building the site",
	Check:   "Checking pages and links",
	Publish: "Putting it live",
}

func (a Action) Wording() string {
	if said, found := wording[a]; found {
		return said
	}
	return string(a)
}

// RunsSiteCode says whether an action carries out something the site wrote.
// Only those are sandboxed; publishing is ChatCMS acting on the project's own
// settings and needs the network and the deploy key the sandbox withholds.
func (a Action) RunsSiteCode() bool {
	return a == Build || a == Check
}

type Step struct {
	Name    string
	Action  Action
	Run     string
	Timeout time.Duration
}

func (s Step) Title() string {
	if s.Name != "" {
		return s.Name
	}
	return s.Action.Wording()
}

type Workflow struct {
	Trigger Trigger
	Steps   []Step
	File    string
}

func (w Workflow) Does(action Action) bool {
	for _, step := range w.Steps {
		if step.Action == action {
			return true
		}
	}
	return false
}

type Error struct {
	File   string
	Step   int
	Key    string
	Reason string
}

func (e *Error) Error() string {
	where := e.File
	if where == "" {
		where = "a workflow"
	}
	if e.Step > 0 {
		where += fmt.Sprintf(", step %d", e.Step)
	}
	if e.Key != "" {
		where += fmt.Sprintf(", %q", e.Key)
	}
	return where + ": " + e.Reason
}

func listed[T ~string](values []T) string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		out = append(out, string(value))
	}
	sort.Strings(out)
	return strings.Join(out, ", ")
}
