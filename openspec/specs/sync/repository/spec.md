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

### Requirement: A working copy reaches its remote with its project's own key

Reaching a remote SHALL use the key belonging to that project, not one shared
across projects.

#### Scenario: Two projects with two remotes

- **WHEN** two projects each reach their own remote
- **THEN** each uses its own key

#### Scenario: A project with no key

- **WHEN** a project has no key and its remote needs one
- **THEN** the failure says the project has no key yet, rather than reporting
  a refusal from the remote

#### Scenario: The key is not left lying about

- **WHEN** the application stops
- **THEN** no decrypted private key remains on disk

### Requirement: A save that names no file saves everything

A save naming no file SHALL make one version of everything the working copy
holds that is not already saved, which is what importing a site is.

#### Scenario: Importing a site

- **WHEN** files are placed in a working copy and a save names none of them
- **THEN** one version is made containing all of them
- **AND** nothing is left unsaved

#### Scenario: Nothing to save

- **WHEN** a save names no file and nothing has changed
- **THEN** it reports that nothing changed rather than making an empty version

### Requirement: The engine says what differs from the published site

The engine SHALL list the files that differ between the site as people see it
and the site as it is now, saying for each whether it was added, edited or
removed, and who last changed it and when.

#### Scenario: a site with unpublished work

- **WHEN** files have been saved since the last publish
- **THEN** each SHALL be listed with what happened to it

#### Scenario: a file that was added and then removed

- **WHEN** a file was added and removed again before publishing
- **THEN** it SHALL NOT be listed, because the published site never had it and
  does not have it now

#### Scenario: a removed file is still attributed

- **WHEN** a file was removed
- **THEN** the person who removed it SHALL be who it is attributed to

#### Scenario: a site that has never been published

- **WHEN** nothing has ever been published
- **THEN** every file SHALL be listed as new

#### Scenario: nothing waiting

- **WHEN** everything saved has been published
- **THEN** nothing SHALL be listed

### Requirement: The engine reads a file as the published site has it

The engine SHALL read a file as the published site holds it, and SHALL say so
when the published site does not hold it at all.

#### Scenario: a file the live site has

- **WHEN** a file is read as the published site has it
- **THEN** its content SHALL be what was published, not what is saved now

#### Scenario: a file the live site does not have

- **WHEN** a file that was never published is read that way
- **THEN** the engine SHALL say that the published site does not have it

### Requirement: The engine publishes to a named line

The engine SHALL send the site to a line named by the caller, leaving the line
the site is edited on alone, and SHALL refuse a name that is unusable.

#### Scenario: a line of its own

- **WHEN** the site is published to a named line
- **THEN** that line SHALL hold the site
- **AND** the line the site is edited on SHALL be unchanged

#### Scenario: no name

- **WHEN** no line is named
- **THEN** it SHALL publish the way it did before, to the line the site is
  edited on

#### Scenario: an unusable name

- **WHEN** a line name is empty, looks like an option, or could escape
- **THEN** publishing SHALL be refused

### Requirement: The engine publishes built pages without the source

Built pages SHALL be sent to a line of their own, without making a version of
the site's own files and without touching the working copy.

#### Scenario: what reaches the line

- **WHEN** built pages are published
- **THEN** the line SHALL hold those pages at its root
- **AND** SHALL NOT hold the site's own files

#### Scenario: the working copy is untouched

- **WHEN** built pages are published
- **THEN** the working copy SHALL be exactly as it was
- **AND** no new version of the site's own files SHALL have been made

#### Scenario: publishing pages twice

- **WHEN** built pages are published a second time
- **THEN** the line SHALL carry both as its history
- **AND** SHALL hold the second publish's pages

#### Scenario: a page that is no longer built

- **WHEN** a page that was published before is no longer produced
- **THEN** it SHALL NOT be on the line any more

#### Scenario: built pages over the site's own line

- **WHEN** built pages would be published to the line the site is edited on
- **THEN** it SHALL be refused
- **AND** the refusal SHALL say that it would replace the site's own files

#### Scenario: nothing to publish

- **WHEN** there are no pages to publish
- **THEN** it SHALL be refused rather than publishing an empty site

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
