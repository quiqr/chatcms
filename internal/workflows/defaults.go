package workflows

// Settings is what a project's own publishing choices tell a default
// workflow. The caller translates the project's settings into it, so this
// package does not have to know how a project is stored.
type Settings struct {
	// ChatCMSBuilds says that publishing sends built pages rather than
	// source files, so somebody has to build them.
	ChatCMSBuilds bool
}

func Default(trigger Trigger, settings Settings) Workflow {
	workflow := Workflow{Trigger: trigger}
	if trigger != OnPublish {
		return workflow
	}
	if settings.ChatCMSBuilds {
		workflow.Steps = append(workflow.Steps, Step{Action: Build})
	}
	workflow.Steps = append(workflow.Steps, Step{Action: Publish})
	return workflow
}

// For answers what happens on a trigger: what the site said when it said
// anything, and the default when it did not. A site that declared a workflow
// with no steps gets that, not the default: saying nothing happens is a thing
// a site is allowed to say.
func (s Set) For(trigger Trigger, settings Settings) Workflow {
	if declared, found := s[trigger]; found {
		return declared
	}
	return Default(trigger, settings)
}
