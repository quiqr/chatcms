## Purpose

Defines what happens when a trigger fires: what runs, in what order, what is
contained, and what a person is told when it goes wrong.

## ADDED Requirements

### Requirement: Steps run in order and stop at the first failure

A workflow's steps SHALL run in the order they are written. A step that fails
SHALL stop the workflow; no later step SHALL run.

#### Scenario: every step works

- **WHEN** no step fails
- **THEN** every step SHALL have run, in order
- **AND** the result SHALL say the workflow succeeded

#### Scenario: a step fails

- **WHEN** a step fails
- **THEN** no step after it SHALL run
- **AND** the result SHALL name the step that failed

#### Scenario: what was done before the failure is still reported

- **WHEN** a workflow fails at its third step
- **THEN** the result SHALL carry the outcome of the first two as well

### Requirement: A failed check publishes nothing

Where a workflow's check fails, nothing SHALL be published, and the site as
people see it SHALL be unchanged.

#### Scenario: the check fails

- **WHEN** the site's check step fails during a publish
- **THEN** the publishing step SHALL NOT run
- **AND** the result SHALL say that nothing was published

#### Scenario: the check passes

- **WHEN** the site's check step succeeds
- **THEN** the workflow SHALL carry on to the next step

### Requirement: The site's own code runs contained

A check step SHALL run through the sandbox: against the site's files, with a
fixed toolbox, with no network, under a time limit and a memory limit.

#### Scenario: a check cannot reach the network

- **WHEN** a check tries to reach the network
- **THEN** it SHALL fail to

#### Scenario: a check that will not end is ended

- **WHEN** a check runs past its time limit
- **THEN** it SHALL be ended and the step SHALL be reported as failed

#### Scenario: a check sees the site

- **WHEN** a check reads a file from the site
- **THEN** it SHALL be the file the site holds

#### Scenario: nothing a check wrote escapes

- **WHEN** a check writes files
- **THEN** none of them SHALL appear in the site or in what is published

### Requirement: Publishing is not sandboxed and runs no site code

The publishing step SHALL be carried out by ChatCMS according to the project's
settings. It SHALL NOT run any command a workflow names.

#### Scenario: publishing cannot be made to run a command

- **WHEN** a workflow tries to give the publishing step a command
- **THEN** the workflow SHALL be refused when it is read

#### Scenario: building hands its result to publishing

- **WHEN** a workflow builds the site and then publishes
- **THEN** what was built SHALL be what the publishing step is given

### Requirement: A failure says what a person can do about it

A failed workflow SHALL carry a sentence written for the person who pressed
the button, saying what happened and what to do. The raw output SHALL be
carried separately.

#### Scenario: an editor is told what happened

- **WHEN** a workflow fails
- **THEN** the result SHALL carry a sentence with no git word in it
- **AND** SHALL say that the live site did not change and the edits are safe

#### Scenario: the raw output is kept apart

- **WHEN** a workflow fails
- **THEN** the raw output SHALL be carried separately from that sentence
- **AND** whether somebody may see it SHALL be decided by the policy function
  rather than by this code

#### Scenario: the raw output is bounded

- **WHEN** a step produces more output than is kept
- **THEN** what is kept SHALL say so and SHALL keep the end

### Requirement: A run is reported step by step

The result SHALL carry, for each step that ran, its name, what it did, whether
it succeeded, and how long it took.

#### Scenario: the steps a person is shown

- **WHEN** a workflow has run
- **THEN** each step's name SHALL be the one the site gave it, or the action's
  own wording where the site gave none
