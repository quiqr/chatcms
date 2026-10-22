## Purpose

Defines what a site may say about what happens when it is saved, previewed or
published, and what it is told when it says something wrong.

## ADDED Requirements

### Requirement: Workflows live in the site

Workflow files SHALL be read from `quiqr/workflows/` in the site, as YAML.
Nothing outside that directory SHALL be read as a workflow.

#### Scenario: a site with workflow files

- **WHEN** a site has files under `quiqr/workflows/`
- **THEN** each one SHALL be read as a workflow

#### Scenario: a site with no workflow directory

- **WHEN** a site has no `quiqr/workflows/` directory
- **THEN** that SHALL NOT be an error
- **AND** every trigger SHALL still have a workflow, worked out by default

#### Scenario: files that are not workflows are left alone

- **WHEN** the directory holds a file that is not YAML by its name
- **THEN** it SHALL be ignored rather than read and refused

### Requirement: A workflow names one trigger and its steps

A workflow SHALL name the trigger it answers and a list of steps. The trigger
SHALL be one of save, preview or publish. The list MAY be empty, which means
the trigger does nothing.

#### Scenario: a workflow is read

- **WHEN** a file names a known trigger and a list of steps
- **THEN** it SHALL be read as the workflow for that trigger

#### Scenario: an unknown trigger is refused

- **WHEN** a file names a trigger that is not save, preview or publish
- **THEN** reading SHALL fail
- **AND** the error SHALL name the file and what the known triggers are

#### Scenario: a workflow with no steps is allowed

- **WHEN** a workflow names a trigger and no steps
- **THEN** it SHALL be read
- **AND** that trigger SHALL do nothing rather than fall back to the default

#### Scenario: two workflows for one trigger are refused

- **WHEN** two files name the same trigger
- **THEN** reading SHALL fail
- **AND** the error SHALL name both files

### Requirement: A step names a built-in action

Each step SHALL name one of the built-in actions: building the site, the
site's own check before publishing, or publishing. A step MAY carry a name to
show a person; without one the action's own wording SHALL be used.

#### Scenario: a step names a known action

- **WHEN** a step names a built-in action
- **THEN** it SHALL be read

#### Scenario: an unknown action is refused

- **WHEN** a step names an action that does not exist
- **THEN** reading SHALL fail
- **AND** the error SHALL name the file, which step it is, and the actions
  that do exist

#### Scenario: a step with no action is refused

- **WHEN** a step names no action
- **THEN** reading SHALL fail naming the file and the step

### Requirement: The site's check carries the command to run

A check step SHALL carry the command to run. The other actions SHALL NOT, as
they run no code the site wrote.

#### Scenario: a check without a command is refused

- **WHEN** a check step carries no command
- **THEN** reading SHALL fail naming the file and the step

#### Scenario: a command on an action that runs nothing is refused

- **WHEN** a step that builds or publishes carries a command
- **THEN** reading SHALL fail
- **AND** the error SHALL say that only the check step runs a command

### Requirement: Steps are ordered so that the result can make sense

Publishing SHALL appear at most once and SHALL be the last step. Building
SHALL appear at most once.

#### Scenario: publishing before another step is refused

- **WHEN** a step follows the publishing step
- **THEN** reading SHALL fail
- **AND** the error SHALL say that nothing runs after publishing

#### Scenario: building twice is refused

- **WHEN** a workflow builds the site twice
- **THEN** reading SHALL fail naming the file

### Requirement: Anything unrecognised is a mistake

A key a workflow file does not define SHALL be refused rather than ignored.

#### Scenario: a misspelled key is caught

- **WHEN** a file carries a key that is not part of the format
- **THEN** reading SHALL fail
- **AND** the error SHALL name the file, the step when it is inside one, and
  the key

#### Scenario: a file that is not YAML at all

- **WHEN** a file named as YAML cannot be read as YAML
- **THEN** reading SHALL fail naming the file

### Requirement: A trigger with no workflow gets a default

Where a site declares no workflow for a trigger, a default SHALL be worked out
from the project's own publishing settings.

#### Scenario: a project that publishes its source

- **WHEN** a project publishes source files and declares no publish workflow
- **THEN** the publish trigger SHALL publish, and SHALL NOT build

#### Scenario: a project that publishes built pages

- **WHEN** a project publishes built HTML and declares no publish workflow
- **THEN** the publish trigger SHALL build the site and then publish

#### Scenario: saving and previewing

- **WHEN** a project declares no workflow for saving or previewing
- **THEN** those triggers SHALL do nothing
