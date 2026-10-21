# content/content-navigation Specification

## Purpose
Defines the navigation structure a site's model implies: the pages and
collections a content creator browses, in the order and under the labels the
model asks for.

## Requirements

### Requirement: Navigation is derived from the model

The system SHALL derive the content navigation from the assembled model, not
from a second structure maintained by hand. A change to the model SHALL change
the navigation with no further step.

#### Scenario: Pages and collections appear

- **WHEN** a model declares two singles and three collections
- **THEN** the navigation offers both groups, with those entries

#### Scenario: A collection reports how many items it has

- **WHEN** a collection's folder holds seven content files
- **THEN** the navigation reports seven for that collection

#### Scenario: An empty collection still appears

- **WHEN** a collection's folder is empty or absent
- **THEN** the collection still appears, reporting none

### Requirement: Labels come from the model

Each collection SHALL carry the singular item name the model gives it, so that
the interface can say "New article" rather than "New item".

#### Scenario: A collection with an item title

- **WHEN** a collection declares its item name
- **THEN** the navigation carries that name for use in labels

#### Scenario: A collection without one

- **WHEN** a collection declares no item name
- **THEN** the navigation carries a neutral fallback
- **AND** the fallback is not a git term or a developer term

### Requirement: The model's menu decides the order

When a model declares a menu, the navigation SHALL follow its sections and
order. When it does not, singles SHALL come before collections and each SHALL
keep the order the model declares.

#### Scenario: A declared menu

- **WHEN** a model declares a menu with two sections
- **THEN** the navigation has those sections, in that order, with the entries
  each lists

#### Scenario: A menu naming something that is not there

- **WHEN** a menu names a single or collection the model does not declare
- **THEN** the entry is omitted from the navigation
- **AND** the omission is reported as a model warning, not as a failure

#### Scenario: No menu declared

- **WHEN** a model declares no menu
- **THEN** singles come first, then collections, each in declaration order

### Requirement: Hidden entries stay hidden

An entry the model marks as hidden SHALL be absent from the navigation while
remaining reachable by its key.

#### Scenario: A hidden collection

- **WHEN** a collection is marked hidden
- **THEN** it does not appear in the navigation
- **AND** asking for it by key still returns it
