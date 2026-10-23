## Purpose

Defines what a machine running ChatCMS is, so that deploying is something the
gate checks rather than something somebody remembers doing.

## ADDED Requirements

### Requirement: A machine runs ChatCMS from the module alone

A machine that imports the module and sets the documented options SHALL run
ChatCMS, with no step performed by hand.

#### Scenario: a fresh machine

- **WHEN** a machine with no ChatCMS on it boots with the module enabled
- **THEN** the database and the service SHALL come up
- **AND** the service SHALL answer on its port

#### Scenario: the sign-in page

- **WHEN** the running service is asked for the sign-in page
- **THEN** it SHALL be served

### Requirement: Migrations run from an empty database at start-up

The schema SHALL be created and upgraded by the server itself, before it
listens, with no separate step.

#### Scenario: nothing in the database

- **WHEN** the service starts against a database with nothing in it
- **THEN** the schema SHALL be created before the service listens
- **AND** the version of the schema SHALL be reported as more than nothing

#### Scenario: starting again

- **WHEN** the service is restarted
- **THEN** it SHALL come back
- **AND** the database SHALL be as it was

### Requirement: The master key never reaches the Nix store

The key SHALL stay where the machine's administrator put it, owned by root,
and no copy of its contents SHALL be in the Nix store.

#### Scenario: where the key lives

- **WHEN** the service is running
- **THEN** the key file SHALL still be owned by root and readable by nobody
  else

#### Scenario: the store

- **WHEN** the Nix store is searched for the key's contents
- **THEN** nothing SHALL be found

#### Scenario: the service reads it anyway

- **WHEN** the service needs the key
- **THEN** it SHALL be given it without the file's permissions being widened
  and without the service running as root

### Requirement: The service runs as its own user

The service SHALL run as a user of its own rather than as root.

#### Scenario: who it runs as

- **WHEN** the service is running
- **THEN** it SHALL run as its own user, not as root
- **AND** its state directory SHALL belong to that user

### Requirement: The pinned tools are the ones used

The git, Hugo and sandbox the service uses SHALL be the ones the deployment
pins, never whatever the machine happens to carry.

#### Scenario: what the service is given

- **WHEN** the service is running
- **THEN** the git, Hugo and sandbox it was given SHALL be ones from the Nix
  store rather than whatever the machine happens to have

### Requirement: Deployment is checked by the gate

Deploying SHALL be checked by booting a machine, on every system where a
machine can be booted.

#### Scenario: a change that breaks deploying

- **WHEN** the gate runs on a system that can boot a machine
- **THEN** the virtual machine test SHALL run
- **AND** a change that stops the module working SHALL fail it
