## ADDED Requirements

### Requirement: The frontend check enforces the interface rules

The frontend check SHALL fail when a colour is written by hand in a component,
when a control is defined below its minimum size, or when a banned word
reaches an editor-facing string.

#### Scenario: Each rule fails the check

- **WHEN** any of the three rules is broken
- **THEN** the frontend check fails and names the file and what is wrong

#### Scenario: The rules run with the type check and the tests

- **WHEN** the frontend check runs
- **THEN** the rules run alongside the type check and the unit tests, not as a
  separate step someone has to remember
