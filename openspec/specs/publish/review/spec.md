# publish/review Specification

## Purpose
Defines what somebody is shown before they decide to publish, so that pressing
the button is a decision rather than a leap.

## Requirements

### Requirement: What changed is said as pages, not as files

The review SHALL list what differs from the published site as pages and
sections, using the names the site's model gives them.

#### Scenario: a page the model knows

- **WHEN** a file the model names as a page has changed
- **THEN** it SHALL be listed under the title the model gives it

#### Scenario: a page in a collection

- **WHEN** a file inside a collection has changed
- **THEN** it SHALL be listed under the collection's title
- **AND** named in words rather than as a file name

#### Scenario: a file the model says nothing about

- **WHEN** a file outside the model has changed
- **THEN** it SHALL still be listed, named from its own path
- **AND** SHALL NOT be left out

### Requirement: Each change says what sort it is and who made it

Every change SHALL carry what happened to it, who last changed it and when.

#### Scenario: the sorts of change

- **WHEN** the review is shown
- **THEN** each change SHALL say whether it is a changed page, a new page, a
  removed page, a new image or a removed image

#### Scenario: who and when

- **WHEN** a change is listed
- **THEN** it SHALL carry the name of whoever last changed it and when they
  did

### Requirement: A changed page says which of its fields changed

For a page that was edited, the review SHALL say which fields differ, naming
each one as the model names it.

#### Scenario: fields are named as the model names them

- **WHEN** a field with a title in the model has changed
- **THEN** it SHALL be listed under that title

#### Scenario: a field with no title in the model

- **WHEN** a field the model gives no title has changed
- **THEN** it SHALL be named from its key, turned into words

#### Scenario: a field inside another

- **WHEN** a field nested inside another has changed
- **THEN** its name SHALL carry the name of the one it sits in

#### Scenario: the page's own text

- **WHEN** the text of the page has changed
- **THEN** it SHALL be listed as a field with a name of its own

#### Scenario: a field that did not change is not listed

- **WHEN** a page has changed
- **THEN** only the fields that differ SHALL be listed

### Requirement: The words that changed are shown

Each changed field SHALL carry the comparison of its old and new text, word by
word, saying for each run whether it was kept, removed or added.

#### Scenario: reading the comparison back

- **WHEN** a field's comparison is read
- **THEN** keeping the runs that were kept and removed SHALL give the text as
  the live site has it
- **AND** keeping the runs that were kept and added SHALL give the text as it
  is now

#### Scenario: only what changed is marked

- **WHEN** one word of a sentence changed
- **THEN** only that word SHALL be marked, and the rest kept

#### Scenario: very long text

- **WHEN** the text is longer than can be compared quickly
- **THEN** it SHALL be reported as replaced rather than compared word by word

### Requirement: The review says when the site was last published

The review SHALL say when the site was last published, where it has been.

#### Scenario: a site that has been published before

- **WHEN** the review is shown for a site published before
- **THEN** it SHALL say when that was

#### Scenario: a site that has never been published

- **WHEN** the review is shown for a site never published
- **THEN** it SHALL say nothing about a last publish rather than failing

### Requirement: Looking is allowed to anybody who may read the project

Seeing the review SHALL need only the right to read the project, not the right
to publish.

#### Scenario: a viewer looks

- **WHEN** somebody who may read the project but not publish asks for the
  review
- **THEN** they SHALL be shown it

#### Scenario: somebody with no part in the project

- **WHEN** somebody with no part in the project asks for the review
- **THEN** they SHALL be told there is nothing there, the same answer as for a
  project that does not exist
