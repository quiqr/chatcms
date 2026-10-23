# content/browsing Specification

## Purpose
Defines what a content creator sees of their site before they open anything.

## Requirements

### Requirement: The browser follows the model

What is listed, and in what order, SHALL come from the site's Quiqr model: the
pages it names, then each collection it declares.

#### Scenario: pages and collections

- **WHEN** a site is browsed
- **THEN** the pages the model names SHALL be one group
- **AND** each collection SHALL be a group of its own, with its own title

#### Scenario: a collection's items

- **WHEN** a collection is listed
- **THEN** its items SHALL come from the files in its folder
- **AND** SHALL be in the same order every time

#### Scenario: a hidden page or collection

- **WHEN** the model marks something hidden
- **THEN** it SHALL NOT be listed

#### Scenario: a site with no model

- **WHEN** a site has no model
- **THEN** the answer SHALL say so rather than show an empty site

### Requirement: Each group says what one of its items is called

A collection SHALL carry the singular word for one of its items, so that
making one can be offered in the site's own words.

#### Scenario: a collection with a word of its own

- **WHEN** the model gives a collection a singular name
- **THEN** that word SHALL be carried

#### Scenario: a collection without one

- **WHEN** the model gives none
- **THEN** a plain word SHALL be used, and it SHALL NOT be a developer's word

#### Scenario: the pages group

- **WHEN** the pages group is listed
- **THEN** it SHALL carry no such word, because it cannot grow

### Requirement: An item waiting to be published is marked

Whether an item has a change nobody has published SHALL come from the same
question the publish review asks, so that the two cannot disagree.

#### Scenario: a changed page

- **WHEN** a page has been saved and not published
- **THEN** it SHALL be marked, with who changed it and when

#### Scenario: a page with nothing waiting

- **WHEN** a page has no unpublished change
- **THEN** it SHALL NOT be marked

### Requirement: Browsing needs only the right to read the project

Seeing what a site holds SHALL need only the right to read it.

#### Scenario: a viewer browses

- **WHEN** somebody who may read but not change the site browses it
- **THEN** they SHALL be shown everything
