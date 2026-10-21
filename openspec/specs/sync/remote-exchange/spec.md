# sync/remote-exchange Specification

## Purpose
Defines how versions made in ChatCMS reach a site's remote and how versions
made elsewhere reach the working copy, including what happens when both sides
changed the same thing.

## Requirements

### Requirement: Remote versions arrive without being asked for

The system SHALL retrieve new versions from the remote on a schedule and on
demand, and SHALL integrate them into the working copy without an editor
taking any action.

#### Scenario: A version pushed straight to the remote

- **WHEN** a version appears on the remote that ChatCMS did not make
- **AND** a sync runs
- **THEN** the working copy contains that version's content
- **AND** status reports that the working copy is level with the remote

#### Scenario: Nothing new on the remote

- **WHEN** a sync runs and the remote has not moved
- **THEN** the working copy is unchanged
- **AND** the sync reports that nothing arrived

#### Scenario: The remote cannot be reached

- **WHEN** a sync runs and the remote does not answer
- **THEN** the working copy is unchanged and usable
- **AND** the failure is reported with the remote named and no credential in
  the message

### Requirement: Local versions are replayed onto the remote

When both sides have new versions, local versions SHALL be replayed on top of
the remote's, so history stays a list of edits rather than a graph of merges.

#### Scenario: Both sides moved

- **WHEN** the working copy has two versions the remote lacks
- **AND** the remote has one version the working copy lacks
- **AND** a sync runs without conflict
- **THEN** history shows the remote's version first, then the two local ones
- **AND** no merge version is created

#### Scenario: Only the local side moved

- **WHEN** the working copy has versions the remote lacks and the remote has
  not moved
- **AND** a publish runs
- **THEN** the remote gains exactly those versions

#### Scenario: The remote rejects the push

- **WHEN** a push is refused because the remote moved in the meantime
- **THEN** no version is lost
- **AND** the next sync retrieves the remote's versions and replays the local
  ones on top

### Requirement: A conflict is reported with every version of the file

When the same file changed on both sides, the system SHALL stop the exchange
and report each conflicted path with the content as it was before, as it is
locally, and as it is on the remote.

#### Scenario: The same file changed on both sides

- **WHEN** a file was changed locally and differently on the remote
- **AND** a sync runs
- **THEN** the sync reports a conflict naming that path
- **AND** the report carries the common ancestor's content, the local content
  and the remote content

#### Scenario: Different files changed on both sides

- **WHEN** one file changed locally and a different file changed on the remote
- **AND** a sync runs
- **THEN** no conflict is reported
- **AND** both changes are present afterwards

#### Scenario: No conflict marker reaches a content file

- **WHEN** a conflict is reported
- **THEN** no file an editor can open contains a conflict marker
- **AND** reading the site's content still yields valid files

### Requirement: A conflict is resolved by choosing content

The system SHALL accept, for each conflicted path, the content that should
win, and SHALL then complete the exchange. It SHALL also accept abandoning the
exchange, leaving the working copy as it was before the sync began.

#### Scenario: Keeping the local content

- **WHEN** a conflict is resolved with the local content for every path
- **THEN** the exchange completes
- **AND** the site's content is the local content
- **AND** the remote's other changes are still present

#### Scenario: Keeping the remote content

- **WHEN** a conflict is resolved with the remote content for every path
- **THEN** the exchange completes
- **AND** the site's content is the remote content

#### Scenario: Keeping something else entirely

- **WHEN** a conflict is resolved with content that matches neither side
- **THEN** the exchange completes with that content

#### Scenario: Abandoning the exchange

- **WHEN** a conflict is abandoned instead of resolved
- **THEN** the working copy is exactly as it was before the sync began
- **AND** no version is lost

#### Scenario: Resolving a path that is not conflicted

- **WHEN** a resolution names a path that was not reported as conflicted
- **THEN** the call fails and the exchange is left as it was

### Requirement: An interrupted exchange leaves a recoverable working copy

If the process stops in the middle of an exchange, the working copy SHALL be
left in a state a later run can either complete or abandon, and no version
SHALL be lost.

#### Scenario: Reopening after an interruption

- **WHEN** an exchange is interrupted partway
- **AND** the working copy is opened again
- **THEN** the system reports that an exchange is in progress
- **AND** abandoning it restores the working copy to its state before the
  exchange

#### Scenario: No version is lost

- **WHEN** an exchange is interrupted and then abandoned
- **THEN** every version made locally before the exchange is still in history
