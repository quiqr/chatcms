## ADDED Requirements

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
