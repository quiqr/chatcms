# interface/editor-language Specification

## Purpose
Defines which words may reach a content creator. The product's promise is that
git stays hidden; a single leaked word breaks it, and leaks happen in strings
nobody reviewed.

## Requirements

### Requirement: Version control words never reach a content creator

Text a content creator can read SHALL NOT use the words the design hand-off
bans: commit, branch, push, pull, merge, repository, deploy, build log, HEAD,
SHA, and conflict. The plain-language replacements SHALL be used instead.

#### Scenario: A banned word in a screen

- **WHEN** a string a content creator can read contains a banned word
- **THEN** the frontend check fails, naming the file and the word

#### Scenario: The replacements are used

- **WHEN** the interface refers to saved work, to changes, to publishing, to a
  version, or to going back to one
- **THEN** it uses those words rather than their version control equivalents

#### Scenario: The word is part of a longer, ordinary word

- **WHEN** a string contains a word that merely spells a banned one inside it
- **THEN** the check does not fail

### Requirement: Technical detail is allowed where it is named as such

Text meant only for an owner or an implementer MAY use technical words, and
SHALL be marked so that the check knows which it is.

#### Scenario: An implementer's view

- **WHEN** a file holds text only an owner or implementer sees
- **AND** it is marked as such
- **THEN** the check allows technical words in it

#### Scenario: Marking is explicit

- **WHEN** a file is not marked
- **THEN** it is treated as readable by a content creator

#### Scenario: The marking is visible in review

- **WHEN** a file is marked as implementer-facing
- **THEN** the marking is a line in the file rather than a setting elsewhere

### Requirement: The check can fail

The check SHALL be proven to fail, so that a passing build means the rule
holds rather than that the rule was never applied.

#### Scenario: A deliberate breach

- **WHEN** a banned word is put into an editor-facing string on purpose
- **THEN** the check fails
