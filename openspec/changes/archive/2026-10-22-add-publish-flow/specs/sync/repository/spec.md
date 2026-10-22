## Purpose

The repository engine learns what publishing needs: what differs from the
published site, what the live site holds, and how to send work to a named line.

## ADDED Requirements

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
