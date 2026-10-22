package workflows_test

import (
	"errors"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/quiqr/chatcms/internal/workflows"
)

func TestAWorkflowIsRead(t *testing.T) {
	workflow, err := workflows.Parse("quiqr/workflows/publish.yaml", []byte(`
on: publish
steps:
  - name: Checking pages and links
    action: check
    run: ./scripts/check-links.sh
    timeout: 90s
  - action: build
  - action: publish
`))
	if err != nil {
		t.Fatalf("Parse() error = %v, want nil", err)
	}

	if workflow.Trigger != workflows.OnPublish {
		t.Errorf("Trigger = %q, want %q", workflow.Trigger, workflows.OnPublish)
	}
	if len(workflow.Steps) != 3 {
		t.Fatalf("read %d steps, want 3", len(workflow.Steps))
	}
	if workflow.Steps[0].Run != "./scripts/check-links.sh" {
		t.Errorf("Run = %q", workflow.Steps[0].Run)
	}
	if workflow.Steps[0].Title() != "Checking pages and links" {
		t.Errorf("Title() = %q, want the name the site gave", workflow.Steps[0].Title())
	}
	if workflow.Steps[1].Title() != "Building the site" {
		t.Errorf("Title() = %q, want the action's own wording", workflow.Steps[1].Title())
	}
	if workflow.Steps[0].Timeout.String() != "1m30s" {
		t.Errorf("Timeout = %v, want 90s", workflow.Steps[0].Timeout)
	}
}

func TestAWorkflowWithNoStepsIsAllowed(t *testing.T) {
	workflow, err := workflows.Parse("quiqr/workflows/save.yaml", []byte("on: save\nsteps: []\n"))
	if err != nil {
		t.Fatalf("Parse() error = %v, want nil", err)
	}
	if len(workflow.Steps) != 0 {
		t.Errorf("read %d steps, want none", len(workflow.Steps))
	}
}

func TestAMistakeNamesTheFileAndTheStep(t *testing.T) {
	cases := map[string]struct {
		yaml     string
		wantStep int
		wantKey  string
		wantSaid string
	}{
		"no trigger": {
			yaml:    "steps: []\n",
			wantKey: "on",
		},
		"unknown trigger": {
			yaml:     "on: tuesday\nsteps: []\n",
			wantKey:  "on",
			wantSaid: "publish",
		},
		"no action": {
			yaml:     "on: publish\nsteps:\n  - name: Something\n",
			wantStep: 1,
			wantKey:  "action",
		},
		"unknown action": {
			yaml:     "on: publish\nsteps:\n  - action: deploy\n",
			wantStep: 1,
			wantKey:  "action",
			wantSaid: "check",
		},
		"a check with nothing to run": {
			yaml:     "on: publish\nsteps:\n  - action: check\n",
			wantStep: 1,
			wantKey:  "run",
		},
		"a command on something that runs none": {
			yaml:     "on: publish\nsteps:\n  - action: publish\n    run: rm -rf /\n",
			wantStep: 1,
			wantKey:  "run",
			wantSaid: "ChatCMS",
		},
		"an unreadable timeout": {
			yaml:     "on: publish\nsteps:\n  - action: check\n    run: true\n    timeout: a while\n",
			wantStep: 1,
			wantKey:  "timeout",
		},
		"a timeout of nothing": {
			yaml:     "on: publish\nsteps:\n  - action: check\n    run: true\n    timeout: 0s\n",
			wantStep: 1,
			wantKey:  "timeout",
		},
		"a misspelled key": {
			yaml:     "on: publish\nstep:\n  - action: publish\n",
			wantSaid: "step",
		},
		"a misspelled key inside a step": {
			yaml:     "on: publish\nsteps:\n  - action: check\n    runs: true\n",
			wantSaid: "runs",
		},
		"not YAML at all": {
			yaml:     "on: publish\n  steps: [\n",
			wantSaid: "",
		},
		"publishing before another step": {
			yaml:     "on: publish\nsteps:\n  - action: publish\n  - action: build\n",
			wantStep: 1,
			wantSaid: "after publishing",
		},
		"building twice": {
			yaml:     "on: publish\nsteps:\n  - action: build\n  - action: build\n",
			wantStep: 2,
			wantSaid: "once",
		},
	}

	for name, test := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := workflows.Parse("quiqr/workflows/publish.yaml", []byte(test.yaml))

			var mistake *workflows.Error
			if !errors.As(err, &mistake) {
				t.Fatalf("Parse() error = %v, want a *workflows.Error", err)
			}
			if mistake.File != "quiqr/workflows/publish.yaml" {
				t.Errorf("File = %q, want the file it was in", mistake.File)
			}
			if test.wantStep != 0 && mistake.Step != test.wantStep {
				t.Errorf("Step = %d, want %d", mistake.Step, test.wantStep)
			}
			if test.wantKey != "" && mistake.Key != test.wantKey {
				t.Errorf("Key = %q, want %q", mistake.Key, test.wantKey)
			}
			if test.wantSaid != "" && !strings.Contains(mistake.Error(), test.wantSaid) {
				t.Errorf("error = %q, want it to mention %q", mistake, test.wantSaid)
			}
		})
	}
}

func TestAMistakeReadsAsASentence(t *testing.T) {
	_, err := workflows.Parse("quiqr/workflows/publish.yaml", []byte("on: publish\nsteps:\n  - action: deploy\n"))

	want := `quiqr/workflows/publish.yaml, step 1, "action": "deploy" is not an action; the actions are build, check, publish`
	if err.Error() != want {
		t.Errorf("error = %q,\nwant        %q", err, want)
	}
}

func TestTwoChecksAreAllowed(t *testing.T) {
	_, err := workflows.Parse("quiqr/workflows/publish.yaml", []byte(`
on: publish
steps:
  - action: check
    run: ./scripts/links.sh
  - action: check
    run: ./scripts/spelling.sh
  - action: publish
`))
	if err != nil {
		t.Errorf("Parse() error = %v, want two checks to be allowed", err)
	}
}

func TestADirectoryIsRead(t *testing.T) {
	site := fstest.MapFS{
		"quiqr/workflows/publish.yaml": {Data: []byte("on: publish\nsteps:\n  - action: publish\n")},
		"quiqr/workflows/save.yml":     {Data: []byte("on: save\nsteps: []\n")},
		"quiqr/workflows/README.md":    {Data: []byte("These are the workflows.\n")},
		"quiqr/workflows/.keep":        {Data: []byte("")},
	}

	set, err := workflows.Load(site)
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}

	if len(set) != 2 {
		t.Fatalf("read %d workflows, want the two YAML ones", len(set))
	}
	if _, found := set[workflows.OnPublish]; !found {
		t.Error("the publish workflow was not read")
	}
	if _, found := set[workflows.OnSave]; !found {
		t.Error("the save workflow written as .yml was not read")
	}
}

func TestASiteWithNoWorkflowsIsNotAnError(t *testing.T) {
	set, err := workflows.Load(fstest.MapFS{"hugo.toml": {Data: []byte("title = \"a site\"\n")}})
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}
	if len(set) != 0 {
		t.Errorf("read %d workflows from a site with none", len(set))
	}
}

func TestTwoWorkflowsForOneTriggerNameBothFiles(t *testing.T) {
	site := fstest.MapFS{
		"quiqr/workflows/a-publish.yaml": {Data: []byte("on: publish\nsteps:\n  - action: publish\n")},
		"quiqr/workflows/b-publish.yaml": {Data: []byte("on: publish\nsteps: []\n")},
	}

	_, err := workflows.Load(site)

	var mistake *workflows.Error
	if !errors.As(err, &mistake) {
		t.Fatalf("Load() error = %v, want a *workflows.Error", err)
	}
	for _, file := range []string{"a-publish.yaml", "b-publish.yaml"} {
		if !strings.Contains(mistake.Error(), file) {
			t.Errorf("error = %q, want it to name %s", mistake, file)
		}
	}
}

func TestAMistakeInOneFileStopsTheLot(t *testing.T) {
	site := fstest.MapFS{
		"quiqr/workflows/publish.yaml": {Data: []byte("on: publish\nsteps:\n  - action: nonsense\n")},
		"quiqr/workflows/save.yaml":    {Data: []byte("on: save\nsteps: []\n")},
	}

	if _, err := workflows.Load(site); err == nil {
		t.Fatal("Load() error = nil, want the mistake to be reported")
	}
}

func TestWhatIsKnownIsListedInOnePlace(t *testing.T) {
	for _, trigger := range workflows.Triggers {
		if !trigger.Known() {
			t.Errorf("%q is listed but not known", trigger)
		}
	}
	for _, action := range workflows.Actions {
		if !action.Known() {
			t.Errorf("%q is listed but not known", action)
		}
	}
	if workflows.Trigger("tuesday").Known() || workflows.Action("deploy").Known() {
		t.Error("something unknown reported itself as known")
	}
}

func TestOnlyTheSitesOwnCodeIsSandboxed(t *testing.T) {
	if !workflows.Check.RunsSiteCode() || !workflows.Build.RunsSiteCode() {
		t.Error("a step that runs the site's code says it does not")
	}
	if workflows.Publish.RunsSiteCode() {
		t.Error("publishing says it runs the site's code; it runs none")
	}
}
