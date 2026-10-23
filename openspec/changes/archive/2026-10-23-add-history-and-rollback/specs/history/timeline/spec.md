## Purpose

Defines what a person sees when they look at what happened to their site, in
words they recognise and with no technical vocabulary at all.

## ADDED Requirements

### Requirement: Versions are grouped into stretches of time

The timeline SHALL be divided into periods a person would name, most recent
first, with the most recent version first inside each period.

#### Scenario: the periods

- **WHEN** the timeline is shown
- **THEN** versions SHALL be grouped as today, yesterday, earlier this week,
  earlier this month, a month within this year, or a month and year

#### Scenario: the order

- **WHEN** two versions fall in the same period
- **THEN** the more recent SHALL be listed first

#### Scenario: a site with no versions

- **WHEN** a site has no versions at all
- **THEN** the timeline SHALL be empty rather than an error

### Requirement: Every version has a title somebody can read

A version SHALL be titled by the note somebody wrote when publishing it. Where
there is none, its own message SHALL be used, and where that message was
written by machinery rather than by a person, a summary of what changed SHALL
be used instead.

#### Scenario: a note somebody wrote

- **WHEN** a publish carried a note
- **THEN** the version it put live SHALL be titled with that note

#### Scenario: no note

- **WHEN** a version carries no note but has a readable message of its own
- **THEN** that message SHALL be the title

#### Scenario: a message no person wrote

- **WHEN** a version's message is machinery's own wording
- **THEN** the title SHALL instead say what changed, counting pages, images
  and other files

#### Scenario: a version that touched nothing

- **WHEN** a version changed no files
- **THEN** it SHALL still have a title rather than an empty one

### Requirement: No version identifier is ever shown

A version's identifier SHALL be carried so that a request can name it, and
SHALL NOT appear in anything a person reads.

#### Scenario: nothing a person reads carries one

- **WHEN** the timeline is shown
- **THEN** no title, name, period or note SHALL contain a version identifier

#### Scenario: the identifier is still there to act on

- **WHEN** a version is listed
- **THEN** it SHALL carry its identifier, so that going back to it is possible

### Requirement: The version the live site holds is marked

Exactly one version SHALL be marked as the one the published site holds, where
anything has been published at all.

#### Scenario: a site with published and unpublished work

- **WHEN** some versions are published and some are waiting
- **THEN** exactly the most recent published version SHALL be marked as the
  one the live site holds

#### Scenario: a site that has never been published

- **WHEN** nothing has been published
- **THEN** no version SHALL be marked as live

### Requirement: Unpublished work is counted

The timeline SHALL say how many versions are waiting to be published.

#### Scenario: work waiting to go out

- **WHEN** versions are waiting to be published
- **THEN** the timeline SHALL say how many

### Requirement: What a version contains can be shown

A version's contents SHALL be available as pages and fields, in the same shape
as the review before publishing.

#### Scenario: asking what a version changed

- **WHEN** one version's contents are asked for
- **THEN** each changed page SHALL be named as the model names it
- **AND** the fields that changed in it SHALL be named and compared

#### Scenario: a version that is not there

- **WHEN** the contents of a version the site does not have are asked for
- **THEN** the answer SHALL be that there is nothing there

### Requirement: The site as it was can be read

The site as one version left it SHALL be readable as a filesystem, without
disturbing the working copy.

#### Scenario: reading an old version

- **WHEN** the site at an earlier version is read
- **THEN** it SHALL hold what that version held
- **AND** the working copy SHALL be unchanged

#### Scenario: walking it

- **WHEN** the site at a version is walked like any other filesystem
- **THEN** every file SHALL be found, with its size and whether it is runnable

### Requirement: Looking needs only the right to read the project

Seeing the history SHALL need only the right to read the project, not the
right to change it.

#### Scenario: a viewer looks

- **WHEN** somebody who may read the project but not change it asks for the
  timeline
- **THEN** they SHALL be shown it

#### Scenario: somebody with no part in the project

- **WHEN** somebody with no part in the project asks
- **THEN** they SHALL be told there is nothing there
