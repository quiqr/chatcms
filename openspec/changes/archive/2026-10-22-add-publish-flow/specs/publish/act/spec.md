## Purpose

Defines what publishing does, what is remembered about it, and what a person
is told when it does not work.

## ADDED Requirements

### Requirement: Publishing runs the site's publish workflow

Publishing SHALL run the workflow for the publish trigger, which is what the
site declared or the default worked out from its settings.

#### Scenario: a site that declares a workflow

- **WHEN** a site declares a publish workflow
- **THEN** that is what runs

#### Scenario: a site that declares none

- **WHEN** a site declares no publish workflow
- **THEN** the default for its publishing settings runs

### Requirement: What is published is what the settings say

A project SHALL publish either its own files or the pages ChatCMS built, to
the line its settings name.

#### Scenario: a project that publishes its source

- **WHEN** a project set to publish source files publishes
- **THEN** the site's own files SHALL reach the named line

#### Scenario: a project that publishes built pages

- **WHEN** a project set to publish built pages publishes
- **THEN** the pages built from the site SHALL reach the named line
- **AND** the site's own files SHALL NOT be on that line

#### Scenario: built pages with nothing built

- **WHEN** a project set to publish built pages publishes and no build ran
- **THEN** publishing SHALL be refused rather than sending the source instead

### Requirement: Publishing does not interrupt editing

Somebody SHALL be able to go on editing while a publish is running, and
neither the publish nor the edit SHALL be spoilt by the other.

#### Scenario: an edit made during a publish

- **WHEN** somebody saves a page while a publish is running
- **THEN** the publish SHALL finish
- **AND** the edit SHALL still be in the site afterwards

#### Scenario: only one publish at a time

- **WHEN** a publish is asked for while one is already running for the same
  project
- **THEN** the second SHALL be refused with a reason worth showing

### Requirement: A failed publish leaves the live site untouched

Where any step of the publish fails, nothing SHALL reach the remote and the
live site SHALL be exactly as it was.

#### Scenario: a step of the workflow fails

- **WHEN** any step before publishing fails
- **THEN** nothing SHALL reach the remote
- **AND** the live site SHALL be exactly as it was

#### Scenario: what a person is told

- **WHEN** a publish fails
- **THEN** the answer SHALL carry a sentence saying that nothing changed on
  the live site and the edits are safe
- **AND** that sentence SHALL carry no git word

### Requirement: The raw output is only for those allowed technical detail

The raw output of a failed publish SHALL be carried apart from the sentence
written for whoever pressed the button, and SHALL only be given to somebody
the policy function allows technical detail.

#### Scenario: an editor reads a failure

- **WHEN** somebody who may not see technical detail reads a failed publish
- **THEN** they SHALL be given the sentence and no raw output

#### Scenario: an implementer reads the same failure

- **WHEN** somebody who may see technical detail reads it
- **THEN** they SHALL be given the raw output as well

### Requirement: Every publish is recorded

Each publish SHALL be recorded with who asked for it, the note they wrote, how
many changes it carried, when it started and finished, how it went, and its
steps.

#### Scenario: the note is kept

- **WHEN** somebody writes a note with a publish
- **THEN** it SHALL be kept with the record

#### Scenario: the history

- **WHEN** a project's publishes are listed
- **THEN** they SHALL be ordered most recent first
- **AND** SHALL carry only that project's publishes

#### Scenario: a publish left running when the server stopped

- **WHEN** the server stops while a publish is recorded as running
- **THEN** on the next start that record SHALL NOT stay running forever

### Requirement: Publishing nothing is not a failure

Publishing a site with nothing waiting SHALL succeed rather than be refused.

#### Scenario: nothing is waiting

- **WHEN** somebody publishes a site with nothing waiting
- **THEN** it SHALL succeed
- **AND** the record SHALL say it carried no changes

### Requirement: Only somebody allowed to publish may

Whether somebody may publish SHALL be decided by the policy function, and a
refusal SHALL mean nothing reaches the remote.

#### Scenario: a viewer tries

- **WHEN** somebody who may read but not publish asks to publish
- **THEN** they SHALL be refused
- **AND** nothing SHALL reach the remote
