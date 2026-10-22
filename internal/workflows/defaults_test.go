package workflows_test

import (
	"testing"
	"testing/fstest"

	"github.com/quiqr/chatcms/internal/workflows"
)

func TestAProjectThatPublishesItsSourceDoesNotBuild(t *testing.T) {
	workflow := workflows.Default(workflows.OnPublish, workflows.Settings{ChatCMSBuilds: false})

	if workflow.Does(workflows.Build) {
		t.Error("a project whose own pipeline builds is built here as well")
	}
	if !workflow.Does(workflows.Publish) {
		t.Error("the default publish workflow does not publish")
	}
	if len(workflow.Steps) != 1 {
		t.Errorf("the default has %d steps, want just the one", len(workflow.Steps))
	}
}

func TestAProjectThatPublishesBuiltPagesBuildsFirst(t *testing.T) {
	workflow := workflows.Default(workflows.OnPublish, workflows.Settings{ChatCMSBuilds: true})

	if len(workflow.Steps) != 2 {
		t.Fatalf("the default has %d steps, want build then publish", len(workflow.Steps))
	}
	if workflow.Steps[0].Action != workflows.Build || workflow.Steps[1].Action != workflows.Publish {
		t.Errorf("the default is %v, want build then publish", workflow.Steps)
	}
}

func TestSavingAndPreviewingDoNothingByDefault(t *testing.T) {
	for _, trigger := range []workflows.Trigger{workflows.OnSave, workflows.OnPreview} {
		for _, builds := range []bool{false, true} {
			workflow := workflows.Default(trigger, workflows.Settings{ChatCMSBuilds: builds})
			if len(workflow.Steps) != 0 {
				t.Errorf("the default for %q has %d steps, want none", trigger, len(workflow.Steps))
			}
		}
	}
}

func TestWhatTheSiteSaidWinsOverTheDefault(t *testing.T) {
	site := fstest.MapFS{
		"quiqr/workflows/publish.yaml": {Data: []byte("on: publish\nsteps:\n  - action: check\n    run: ./check.sh\n  - action: publish\n")},
	}
	set, err := workflows.Load(site)
	if err != nil {
		t.Fatal(err)
	}

	workflow := set.For(workflows.OnPublish, workflows.Settings{ChatCMSBuilds: true})

	if workflow.Does(workflows.Build) {
		t.Error("the default was used even though the site said what happens")
	}
	if !workflow.Does(workflows.Check) {
		t.Error("what the site said was not used")
	}
}

func TestASiteCanSayThatNothingHappens(t *testing.T) {
	site := fstest.MapFS{
		"quiqr/workflows/publish.yaml": {Data: []byte("on: publish\nsteps: []\n")},
	}
	set, err := workflows.Load(site)
	if err != nil {
		t.Fatal(err)
	}

	workflow := set.For(workflows.OnPublish, workflows.Settings{ChatCMSBuilds: true})

	if len(workflow.Steps) != 0 {
		t.Errorf("a site that said nothing happens got %d steps", len(workflow.Steps))
	}
}

func TestATriggerTheSiteSaidNothingAboutGetsTheDefault(t *testing.T) {
	site := fstest.MapFS{
		"quiqr/workflows/save.yaml": {Data: []byte("on: save\nsteps: []\n")},
	}
	set, err := workflows.Load(site)
	if err != nil {
		t.Fatal(err)
	}

	workflow := set.For(workflows.OnPublish, workflows.Settings{ChatCMSBuilds: true})

	if len(workflow.Steps) != 2 {
		t.Errorf("the publish trigger got %d steps, want the default two", len(workflow.Steps))
	}
}
