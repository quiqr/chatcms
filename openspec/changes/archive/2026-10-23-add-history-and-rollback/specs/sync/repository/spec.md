## Purpose

The repository engine learns what the history needs: a version that holds an
earlier version's content, what one version changed, and the site as it was.

## ADDED Requirements

### Requirement: The engine makes a version holding an earlier one's content

The engine SHALL make a new version whose content is an earlier version's,
without removing any version from the history.

#### Scenario: the content comes back

- **WHEN** a version is made from an earlier version's content
- **THEN** the working copy SHALL hold what that earlier version held
- **AND** files the earlier version did not have SHALL be gone from it

#### Scenario: the new version is attributed

- **WHEN** such a version is made
- **THEN** it SHALL be attributed to whoever asked for it

#### Scenario: nothing would change

- **WHEN** the site already holds exactly that content
- **THEN** the engine SHALL refuse rather than make an empty version

#### Scenario: during an exchange with the remote

- **WHEN** an exchange with the remote is unfinished
- **THEN** the engine SHALL refuse

### Requirement: The engine says what one version changed

The engine SHALL list the files one version added, edited or removed, compared
with the version before it.

#### Scenario: a version in the middle of the history

- **WHEN** one version's changes are asked for
- **THEN** each file it added, edited or removed SHALL be listed
- **AND** each SHALL be attributed to whoever made that version

#### Scenario: the first version of a site

- **WHEN** the first version's changes are asked for
- **THEN** every file in it SHALL be listed as new

### Requirement: The engine reads the site as it was

The engine SHALL give the site at a version, and the site just before a
version, as read-only filesystems.

#### Scenario: at a version

- **WHEN** the site at a version is read
- **THEN** it SHALL hold that version's files and no later ones
- **AND** the working copy SHALL be undisturbed

#### Scenario: before a version

- **WHEN** the site just before a version is read
- **THEN** it SHALL hold what the files said before that version changed them

#### Scenario: before the first version

- **WHEN** the site before the first version is read
- **THEN** it SHALL be empty rather than an error

#### Scenario: it behaves as a filesystem

- **WHEN** the site at a version is walked, listed and read
- **THEN** it SHALL behave as any other read-only filesystem does, with each
  file's size and whether it is runnable reported the same way by a listing
  and by the file itself

### Requirement: A version identifier from outside is checked

An identifier that did not come from this engine SHALL be refused rather than
passed to git.

#### Scenario: anything that is not an identifier

- **WHEN** an identifier is empty, names a reference, or carries anything but
  the characters an identifier is made of
- **THEN** every method taking one SHALL refuse it

#### Scenario: an identifier of the right shape for a version that is absent

- **WHEN** an identifier is well formed but the site has no such version
- **THEN** the answer SHALL be that there is no such version
