## ADDED Requirements

### Requirement: The platform admin has an area of its own

Administering the installation SHALL happen in an area that says what it is,
with its own shell, so that nobody is in it by accident.

#### Scenario: Where you are

- **WHEN** a platform admin is in the area
- **THEN** the shell SHALL say so, and SHALL offer the way back to their own
  projects

#### Scenario: Anybody else

- **WHEN** somebody who does not administer the installation asks for the
  area
- **THEN** it SHALL be refused, and nothing SHALL link to it from their
  screens

### Requirement: Every organisation is listed with what it holds

The list of organisations SHALL say, for each one, who administers it, what
it holds against its limits, and when it was made. Searching and filtering
SHALL be answered by the server.

#### Scenario: One with nobody to administer it

- **WHEN** an organisation has no admin
- **THEN** the list SHALL say so, in the colour the design gives a warning

#### Scenario: Filtering

- **WHEN** a platform admin asks for the organisations at a limit, or those
  with no admin
- **THEN** only those SHALL be listed, decided by the server

#### Scenario: Searching

- **WHEN** a platform admin searches by name or address
- **THEN** only the organisations that match SHALL be listed

### Requirement: Removing an organisation removes what it holds in ChatCMS

Removing an organisation SHALL need administration of the installation, and
SHALL remove its sites and its memberships from ChatCMS. It SHALL NOT touch
anything in git, and SHALL be refused until the organisation's name has been
typed.

This replaces the rule that removing was refused while an organisation still
held sites. The hand-off of 2 November draws the opposite: the dialog lists
what goes and asks for the name, which is the protection the refusal stood
in for.

#### Scenario: What goes

- **WHEN** an organisation is removed
- **THEN** its sites SHALL no longer be in ChatCMS
- **AND** the accounts of its members SHALL remain, having lost their access
  to it

#### Scenario: What stays

- **WHEN** an organisation is removed
- **THEN** every git remote its sites used SHALL be untouched
- **AND** every live website SHALL keep working

#### Scenario: Asked for twice

- **WHEN** the name typed does not match the organisation's
- **THEN** removing SHALL NOT be offered

#### Scenario: Who may

- **WHEN** somebody who does not administer the installation asks to remove
  an organisation
- **THEN** it SHALL be refused

#### Scenario: What is written down

- **WHEN** an organisation is removed
- **THEN** the act SHALL be in the administration record, with how many sites
  went with it

## REMOVED Requirements

### Requirement: An administrator removes an organisation
