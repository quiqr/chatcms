# Spec Delta

## Purpose

Defines what a development machine gets that a deployment never does: people
seeded for every role, and signing in as one of them in a single step.

## ADDED Requirements

### Requirement: Every role has a seeded person who can sign in

The development seed SHALL leave one person per role able to sign in with the
same passphrase: one who administers the installation, an implementer, an
editor and a viewer. The editor and the viewer SHALL hold their role on a
project that has a site.

#### Scenario: The editor signs in after seeding

- **WHEN** the database has been seeded and the editor's address is entered
  with the seeded passphrase
- **THEN** sign-in SHALL succeed
- **AND** the editor SHALL see the project with a site with the editor role

#### Scenario: The viewer signs in after seeding

- **WHEN** the database has been seeded and the viewer's address is entered
  with the seeded passphrase
- **THEN** sign-in SHALL succeed
- **AND** the viewer SHALL see that project with the viewer role

### Requirement: One seeded invitation stays open

The seed SHALL also invite one person who does not accept, so that an open
invitation can still be looked at on a development machine and in the
acceptance suite.

#### Scenario: The open invitation after seeding

- **WHEN** the database has been seeded
- **THEN** one invited person on the project with a site SHALL still be to
  accept
- **AND** signing in as that person SHALL be refused

### Requirement: The development server offers sign-in as each role

The sign-in screen SHALL, when served by the frontend development server, show
a block marked as development with one button per seeded role. A button SHALL
sign in as that person through the ordinary password sign-in.

#### Scenario: Signing in as the editor in one click

- **WHEN** the sign-in screen is opened from the development server and the
  editor button is pressed
- **THEN** the person SHALL be signed in as the seeded editor
- **AND** the request SHALL be the same password sign-in a typed address and
  passphrase make

#### Scenario: A seeded person who is missing

- **WHEN** a button is pressed and that person does not exist in the database
- **THEN** the screen SHALL show the same refusal as a wrong password

### Requirement: A production build carries nothing of development sign-in

A production build of the frontend SHALL contain neither the development
block nor the seeded passphrase, and the frontend check SHALL fail when it
does.

#### Scenario: The sign-in screen in a deployment

- **WHEN** the sign-in screen is served by a built binary
- **THEN** no development block SHALL be shown

#### Scenario: The passphrase leaks into the bundle

- **WHEN** the production bundle contains the seeded passphrase
- **THEN** the frontend check SHALL fail
