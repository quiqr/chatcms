# interface/app-shell Specification

## Purpose
Defines the frames every screen sits in, and how each behaves as the window
narrows, so that working on a phone is the same product rather than a reduced
one.

## Requirements

### Requirement: A screen inside a project has the project frame

A screen inside a project SHALL sit in a frame carrying the project's
navigation, the current sync state, and the publish action with a count of
what is unpublished.

#### Scenario: The frame carries navigation

- **WHEN** a screen inside a project is shown
- **THEN** the frame offers the site's sections and, for an owner or
  implementer, the setup sections

#### Scenario: Setup is hidden, not disabled

- **WHEN** the person is an editor or a viewer
- **THEN** the setup sections are absent rather than shown greyed out

#### Scenario: The publish action carries its count

- **WHEN** there are unpublished changes
- **THEN** the publish action shows how many

### Requirement: The frame reflows rather than breaking

The frame SHALL reflow as the window narrows: the navigation gives way to the
content, and below the phone threshold the phone frame is used instead.

#### Scenario: A narrow window

- **WHEN** the window is too narrow for navigation beside content
- **THEN** the navigation moves above the content rather than squeezing it

#### Scenario: A phone

- **WHEN** the window is at or below the phone threshold
- **THEN** the phone frame is used, with its bottom tab bar
- **AND** the project navigation frame is not shown as well

#### Scenario: No sideways scrolling

- **WHEN** any screen is shown at the phone width
- **THEN** the page does not scroll sideways

### Requirement: The phone frame respects the device

The phone frame SHALL keep its actions reachable and clear of the device's own
furniture.

#### Scenario: The bottom bar clears the device

- **WHEN** the phone frame is shown on a device with a home indicator
- **THEN** the bottom bar is padded clear of it

#### Scenario: The editor swaps the tab bar for its actions

- **WHEN** a page is being edited on a phone
- **THEN** the tab bar gives way to the editing actions, including publish

### Requirement: Organisation screens have their own frame

A screen above the level of a single project SHALL use a header rather than
the project navigation.

#### Scenario: The project list

- **WHEN** the list of projects is shown
- **THEN** it sits under a header with the organisation switcher, not inside a
  project's navigation

### Requirement: The editor frame leaves room for what comes later

The editor SHALL lay out as a row of slots with a narrow icon rail, so that a
third panel can be added without the layout being rebuilt.

#### Scenario: The editor has its slots

- **WHEN** the editor frame is rendered
- **THEN** it has a slot for the form, a slot for the preview, and a slot that
  is empty for now

#### Scenario: The rail replaces the full navigation

- **WHEN** the editor is shown
- **THEN** the project navigation is a narrow rail rather than the full
  sidebar

### Requirement: Every screen of this milestone is reachable

Each screen the milestone names SHALL have an address, so that a screen can be
built without navigation being invented alongside it.

#### Scenario: A screen has an address

- **WHEN** any milestone screen is named
- **THEN** there is an address that reaches it

#### Scenario: An address that reaches nothing

- **WHEN** an address is used that names no screen
- **THEN** the person is told so in plain language and offered a way back
