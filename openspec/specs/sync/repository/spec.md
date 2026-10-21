# sync/repository Specification

## Purpose
Defines how a site's working copy is created, read, written and described, and
what counts as a version, so that the rest of the system never has to know
that git is underneath.

## Requirements

### Requirement: Only one component runs git

Every git invocation in the system SHALL originate from the repository
component. No other package SHALL execute the git binary.

#### Scenario: The rule is enforced, not documented

- **WHEN** any package outside the repository component executes git
- **THEN** the test suite fails and names that package

#### Scenario: The binary is the configured one

- **WHEN** the system runs git
- **THEN** it runs the binary named by configuration
- **AND** it does not search the executable path

### Requirement: A working copy is opened or created, never guessed

The system SHALL create a working copy for a site by cloning its remote, and
SHALL open an existing working copy without touching the remote. Opening a
directory that is not a working copy SHALL fail rather than create one.

#### Scenario: A first clone

- **WHEN** a site is cloned into an empty directory
- **THEN** the directory holds the site's content at the remote's current state
- **AND** the working copy reports the remote it came from

#### Scenario: Opening an existing working copy

- **WHEN** an existing working copy is opened
- **THEN** no network request is made
- **AND** the content on disk is untouched

#### Scenario: Opening something that is not a working copy

- **WHEN** a directory holding ordinary files is opened as a working copy
- **THEN** the call fails with an error saying so
- **AND** the directory is unchanged

#### Scenario: Cloning where a working copy already exists

- **WHEN** a clone is asked for in a directory that already holds one
- **THEN** the call fails rather than discarding what is there

### Requirement: Content is read through a filesystem

The working copy SHALL expose its content as a read-only filesystem, so that
components which only read content need no knowledge of version control.

#### Scenario: Reading a content file

- **WHEN** a component asks the working copy for its filesystem
- **AND** reads a path that exists in the site
- **THEN** it gets the file's current content

#### Scenario: The version control directory is not content

- **WHEN** a component lists the working copy's filesystem
- **THEN** version control's own bookkeeping directory is not among the entries

### Requirement: Saving produces a version attributed to a person

Writing one or more files SHALL produce exactly one version carrying the
message given, attributed to the person who made the edit, with the system
recorded as the agent that wrote it.

#### Scenario: A save becomes one version

- **WHEN** two files are written in one save
- **THEN** the history gains exactly one version containing both changes

#### Scenario: The version is attributed to the editor

- **WHEN** a save is made on behalf of a named person with an email address
- **THEN** the version records that person as its author

#### Scenario: A save that changes nothing

- **WHEN** a save writes content identical to what is already there
- **THEN** no version is created
- **AND** the call reports that nothing changed rather than failing

#### Scenario: A save that deletes

- **WHEN** a save removes a file
- **THEN** the version records the removal
- **AND** the file is gone from the working copy's filesystem

#### Scenario: A path outside the working copy

- **WHEN** a save names a path that escapes the working copy
- **THEN** the call fails and writes nothing

### Requirement: History is a list of versions in plain terms

The system SHALL report history as versions, each carrying an identifier, a
time, an author, a message and the paths it changed.

#### Scenario: Listing recent versions

- **WHEN** history is asked for with a limit
- **THEN** it returns at most that many versions, newest first

#### Scenario: A version names what it changed

- **WHEN** a version is read
- **THEN** it lists the content paths that changed in it

#### Scenario: An empty site

- **WHEN** history is asked for on a working copy with no versions
- **THEN** it returns an empty list rather than failing

### Requirement: Status is computed, never remembered

The system SHALL answer whether there is unsaved work, how many versions are
not yet on the remote, and whether the remote is known to have moved, by
inspecting the working copy rather than by reading a value it stored earlier.

#### Scenario: A clean working copy

- **WHEN** status is asked for after a save
- **THEN** it reports no unsaved work

#### Scenario: Versions not yet published

- **WHEN** three versions were made since the last exchange with the remote
- **THEN** status reports three unpublished versions

#### Scenario: Status after the state file is discarded

- **WHEN** any cached state the system holds is deleted
- **AND** status is asked for again
- **THEN** it reports the same answer

### Requirement: One writer at a time per site

Writes to one working copy SHALL be serialised, so two concurrent saves cannot
interleave into one version or corrupt the index.

#### Scenario: Two concurrent saves

- **WHEN** two saves to the same site run at the same time
- **THEN** both succeed
- **AND** the history gains exactly two versions
- **AND** neither version contains the other's files

#### Scenario: Different sites do not block each other

- **WHEN** saves to two different sites run at the same time
- **THEN** neither waits for the other
