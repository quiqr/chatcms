# acceptance/reference-instance Specification

## Purpose
Defines what a reference instance of the old Quiqr is: an installation of the
application whose behaviour ChatCMS promises to reproduce, set up by one
command, pinned so that an answer it gives today is the answer it gave
yesterday, and kept out of the gate because it reaches the network.

## Requirements

### Requirement: One command sets the reference up

A single command SHALL bring the old Quiqr up on a machine that has none of
it, with a site mounted, and SHALL end by saying where it is and how to sign
in. Nothing in the set-up SHALL require a browser, a display or a click.

#### Scenario: a machine with nothing

- **WHEN** the command runs on a machine with no checkout, no node modules
  and no configuration of the old application
- **THEN** it SHALL fetch what it needs, build it, start it, and print an
  address, a user and a password
- **AND** the site it mounts SHALL be editable through that address

#### Scenario: no display

- **WHEN** the command runs with no graphical session
- **THEN** it SHALL still succeed, because the edition it starts serves its
  interface over HTTP rather than drawing it

#### Scenario: a second run

- **WHEN** the command runs again with the state already in place
- **THEN** it SHALL NOT fetch or rebuild anything
- **AND** it SHALL reach the same address with the same site

### Requirement: Everything that decides behaviour is pinned and reported

Every version that can change what the reference answers SHALL be pinned in
one place, and SHALL be printed when the reference starts. A reference that
cannot say what it is cannot be cited.

#### Scenario: what is printed

- **WHEN** the reference starts
- **THEN** it SHALL print the commit of the application, the commit of the
  site it mounts, and the versions of the programs it runs them with

#### Scenario: a version moves

- **WHEN** a pinned version is changed
- **THEN** the change SHALL be one edit in one file
- **AND** any recorded answer that changes because of it SHALL be traceable
  to that edit

#### Scenario: nothing floats

- **WHEN** the reference is set up twice from the same pins
- **THEN** both SHALL run the same code, whatever the day

### Requirement: The reference reaches outside, and is therefore not in the gate

The reference SHALL be run by hand and SHALL NOT be part of `nix flake check`,
because it needs the network and a package manager. Anything recorded from it
SHALL be committed, so that the tests which cite it keep running offline.

#### Scenario: the gate stays offline

- **WHEN** `nix flake check` runs
- **THEN** it SHALL NOT fetch the old application, and SHALL NOT need the
  reference to be running

#### Scenario: a recorded answer

- **WHEN** an answer is taken from the reference
- **THEN** the answer SHALL be committed rather than fetched again at test
  time

### Requirement: The reference owns its state and gives it back

Everything the reference writes SHALL live under one directory that is not
part of the repository's tracked files. Stopping SHALL leave nothing running,
and starting fresh SHALL leave nothing behind.

#### Scenario: stopping

- **WHEN** the reference is stopped
- **THEN** no process it started SHALL still be running, and no port it took
  SHALL still be held

#### Scenario: starting fresh

- **WHEN** a fresh set-up is asked for
- **THEN** the checkout, the installed packages, the application's
  configuration and its mounted site SHALL all be thrown away first

#### Scenario: the repository stays clean

- **WHEN** the reference has been run
- **THEN** the repository SHALL show no change the person running it did not
  make themselves

### Requirement: The mounted site exercises the fields

The site the reference mounts SHALL declare the field types whose behaviour
is in question, so that the reference shows them rather than being an empty
application. The reference SHALL assert nothing about any of them.

#### Scenario: the field types

- **WHEN** the mounted site's model is read
- **THEN** it SHALL declare every field type that the project's own form
  draws, and every type the project draws as a plain text box for want of
  anything better

#### Scenario: a type that is not covered

- **WHEN** a field type the old application draws is absent from the mounted
  site
- **THEN** that absence SHALL be written down where somebody looking for the
  type will find it, rather than left to be discovered

#### Scenario: no assertions

- **WHEN** the reference is set up
- **THEN** no test SHALL pass or fail as a result, because setting the
  reference up is not a claim about anything

### Requirement: An answer can be taken out of the reference and kept

The reference SHALL offer a way to record what it did to a file, as files
rather than as a description: the model, the content before, the content
after, and the edit that was made. A recorded answer SHALL be readable by a
test that runs with the reference not running and the network away.

#### Scenario: recording one edit

- **WHEN** a value is changed through the old application's own form and
  saved
- **THEN** the reference SHALL be able to produce the file before, the file
  after, and the edit in a form a test can replay

#### Scenario: what a recorded answer carries

- **WHEN** a recorded answer is read later
- **THEN** it SHALL say which versions produced it, so that an answer and the
  application that gave it cannot drift apart unnoticed

#### Scenario: recording nothing by accident

- **WHEN** the application wrote more than the edit asked for
- **THEN** that SHALL be visible in the recorded answer rather than smoothed
  away, because it is exactly the kind of thing the reference exists to find
