## Purpose

Defines when a preview starts, when it stops, and how many may run.

## ADDED Requirements

### Requirement: A change shows without the preview being restarted

A change saved to a site SHALL show in a running preview within seconds,
without the preview being started again.

#### Scenario: a saved page

- **WHEN** a page is saved while a preview is running
- **THEN** the change SHALL show in the preview within seconds
- **AND** the same preview SHALL still be the one running

### Requirement: Asking twice uses the preview already running

One site SHALL have at most one preview running.

#### Scenario: a second request

- **WHEN** a preview of the same site is asked for again
- **THEN** the one already running SHALL be used

### Requirement: A preview nobody is looking at is stopped

A preview SHALL be stopped when nobody has asked for it for a stated length of
time.

#### Scenario: nobody looking

- **WHEN** nobody has asked for a preview for longer than that
- **THEN** it SHALL be stopped and its resources released

#### Scenario: somebody looking

- **WHEN** somebody asks for it again
- **THEN** it SHALL be kept, and SHALL still answer

### Requirement: Only so many previews run at once

The number of previews running at once SHALL be capped, so that one busy
afternoon cannot exhaust the machine.

#### Scenario: the cap

- **WHEN** as many previews are running as are allowed
- **THEN** another SHALL be refused with a reason worth showing
- **AND** the ones running SHALL be untouched

### Requirement: A preview whose site moved is started again

A preview serving a working copy that is no longer where it was SHALL be
stopped and started afresh.

#### Scenario: a working copy made afresh

- **WHEN** a preview is asked for a site whose working copy is somewhere else
- **THEN** the old preview SHALL be stopped and a new one started
- **AND** only one SHALL be running for that site

### Requirement: Nothing outlives the app

Every preview SHALL be stopped when the app stops, and stopping SHALL be safe
to ask for more than once.

#### Scenario: the app stops

- **WHEN** the app stops
- **THEN** every preview SHALL be stopped

#### Scenario: stopping twice

- **WHEN** stopping is asked for again
- **THEN** it SHALL do nothing rather than fail
