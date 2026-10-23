# acceptance/coverage Specification

## Purpose
Defines which acceptance tests exist, what each one proves, and what every
phone pass adds on top.

## Requirements

### Requirement: Every acceptance test from the briefing is automated

The suite SHALL hold one spec for each of the nine acceptance tests in
`docs/BRIEFING-M01.md` section 6, and each SHALL drive the product the way a
person would: through the browser, against the built binary.

#### Scenario: an organisation and an invitation

- **WHEN** an owner makes an organisation and invites somebody as an editor
- **THEN** the invitation SHALL be usable to sign in
- **AND** the invited person SHALL see the project and not the owner's screens

#### Scenario: importing a site

- **WHEN** a site is imported by address
- **THEN** its pages SHALL be listed
- **AND** the key it offers SHALL be shown for putting on the remote

#### Scenario: the forms the model describes

- **WHEN** a page is opened whose model names each field type
- **THEN** every field SHALL be on the screen with its own label

#### Scenario: saving keeps the file the way it was

- **WHEN** a field is changed and saved
- **THEN** the file SHALL differ only in the line that was edited

#### Scenario: a page and a picture

- **WHEN** a page is made, edited and removed, and a picture uploaded
- **THEN** each SHALL be visible on the remote afterwards

#### Scenario: the preview follows the form

- **WHEN** a field is changed and saved
- **THEN** the preview SHALL show the new words without being reloaded by hand

#### Scenario: work done elsewhere

- **WHEN** a change is pushed straight to the remote
- **THEN** it SHALL appear in the app without anybody asking for it

#### Scenario: publishing

- **WHEN** a publish is asked for
- **THEN** what changed SHALL be shown before anything leaves
- **AND** the remote SHALL hold the change afterwards

#### Scenario: going back

- **WHEN** an earlier version is chosen
- **THEN** the site SHALL read as it did then
- **AND** going back SHALL itself be a version that can be undone

### Requirement: Every spec runs at both widths

Each spec SHALL run once at 1440 px and once at 390 px, in the same suite,
so that the desk and the phone cannot drift apart.

#### Scenario: the phone rules

- **WHEN** a spec finishes its phone pass
- **THEN** the page SHALL NOT scroll sideways
- **AND** nothing that can be pressed SHALL be under 44 px in either direction

#### Scenario: screens no spec lands on

- **WHEN** a screen is reached by none of the other specs
- **THEN** a spec SHALL visit it on a phone and check the same two rules

### Requirement: A failing behaviour fails its spec

A spec SHALL fail when the behaviour it describes is broken, and SHALL NOT
pass by waiting for a clock or by asserting less than it claims.

#### Scenario: the behaviour is broken deliberately

- **WHEN** the product is changed so that one acceptance test's behaviour no
  longer holds
- **THEN** that spec SHALL fail
- **AND** the others SHALL still pass

#### Scenario: no sleeping

- **WHEN** a spec waits for something
- **THEN** it SHALL wait for that thing, with a deadline, and never for a
  fixed stretch of time

### Requirement: Every screen is captured beside its drawing

The suite SHALL be able to capture each screen at both widths, and the
differences from `docs/design/reference/screenshots/` SHALL be written down
rather than silently accepted.

#### Scenario: a capture run

- **WHEN** the capture is run
- **THEN** each screen SHALL be written at 1440 px and at 390 px

#### Scenario: a difference from the drawing

- **WHEN** a screen differs from the drawing it came from
- **THEN** the difference SHALL be listed in `docs/design/COMPARISON.md` with
  what was decided about it
