## Purpose

Defines how a site's standing is said, so that somebody editing always knows
whether their work is safe.

## ADDED Requirements

### Requirement: A site stands in one of five ways

The standing of a site SHALL be one of: everything is saved, something is
being saved, the site could not be reached, somebody else updated it, or two
people changed the same thing.

#### Scenario: everything saved

- **WHEN** nothing is waiting and nothing arrived
- **THEN** the standing SHALL say everything is saved

#### Scenario: work not yet saved

- **WHEN** something has been typed and not yet written
- **THEN** the standing SHALL say it is saving

#### Scenario: an unfinished question

- **WHEN** an exchange with the remote is waiting on an answer
- **THEN** the standing SHALL say two people changed the same thing

### Requirement: Being unreachable is a standing, not a failure

A site that cannot be reached SHALL be reported as a standing of its own,
with the work kept here until it can be.

#### Scenario: the site cannot be reached

- **WHEN** bringing in changes finds the remote unreachable
- **THEN** the standing SHALL say so and that the work is here
- **AND** it SHALL NOT be reported as something going wrong

### Requirement: What arrived says whose it was

Work brought in from other people SHALL name each of them once.

#### Scenario: somebody else's work arrives

- **WHEN** changes from other people are brought in
- **THEN** the standing SHALL say that somebody else updated the site
- **AND** SHALL name each of them once

#### Scenario: nothing arrived

- **WHEN** nothing came in
- **THEN** nobody SHALL be named

### Requirement: How much is waiting is said

The standing SHALL say how much work is waiting to be published.

#### Scenario: unpublished work

- **WHEN** work is waiting to be published
- **THEN** the standing SHALL say how much

### Requirement: Every word a content creator reads is theirs

Everything the standing says SHALL be in the words
`docs/design/DESIGN.md` section 7 allows.

#### Scenario: the words used

- **WHEN** any standing is shown
- **THEN** what it says SHALL carry no version control word

### Requirement: Technical detail is for implementers and nobody else

Where somebody may see technical detail, the standing MAY carry it. Where they
may not, it SHALL be left out of the answer entirely rather than hidden.

#### Scenario: an editor

- **WHEN** somebody who may not see technical detail asks how the site stands
- **THEN** the answer SHALL carry none

#### Scenario: an implementer

- **WHEN** somebody who may asks
- **THEN** the answer SHALL carry where the site keeps its work

### Requirement: Looking needs only the right to read the project

Seeing how a site stands SHALL need only the right to read it.

#### Scenario: a viewer looks

- **WHEN** somebody who may read but not change the site asks how it stands
- **THEN** they SHALL be told
