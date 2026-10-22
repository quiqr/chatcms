package workflows

import (
	"strings"
	"time"
)

type State string

const (
	Succeeded State = "succeeded"
	Failed    State = "failed"
	NotRun    State = "not-run"
)

type StepOutcome struct {
	Title    string
	Action   Action
	State    State
	Duration time.Duration

	// Detail is the raw output of the step. Whether anybody may see it is
	// decided by policy.Can with see-technical-detail, where it is rendered.
	// This package never decides, which is why it is not in Trouble.
	Detail string
}

type Outcome struct {
	Trigger Trigger
	Steps   []StepOutcome

	// Trouble is written for the person who pressed the button. It carries no
	// git word and no raw output, so it is safe to show to anybody who may
	// see the project at all.
	Trouble *Trouble
}

type Trouble struct {
	Step    string
	Because string
	Advice  string
}

const reassurance = "Nothing changed on your live site, and your edits are safe."

var openings = map[Trigger]string{
	OnPublish: "We couldn't publish this time.",
	OnSave:    "Something went wrong after saving.",
	OnPreview: "We couldn't get the preview ready.",
}

// Sentence is what the person who pressed the button is shown. It carries no
// git word and no raw output, which is why the raw output lives on the step
// rather than here.
func (o *Outcome) Sentence() string {
	if o == nil || o.Trouble == nil {
		return ""
	}
	opening, known := openings[o.Trigger]
	if !known {
		opening = "Something went wrong."
	}
	said := []string{opening, reassurance}
	if o.Trouble.Because != "" {
		said = append(said, o.Trouble.Because)
	}
	if o.Trouble.Advice != "" {
		said = append(said, o.Trouble.Advice)
	}
	return strings.Join(said, " ")
}

func (o *Outcome) Succeeded() bool { return o != nil && o.Trouble == nil }

func (o *Outcome) Ran(action Action) bool {
	for _, step := range o.Steps {
		if step.Action == action && step.State == Succeeded {
			return true
		}
	}
	return false
}
