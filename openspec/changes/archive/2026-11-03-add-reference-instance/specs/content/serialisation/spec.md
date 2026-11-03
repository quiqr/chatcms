# Spec Delta

## ADDED Requirements

### Requirement: One recorded case comes from the application, not its libraries

At least one case SHALL be recorded by editing a value in the old Quiqr
itself, through its own form, and keeping the file it wrote. The recorded
case SHALL carry the model, the file before, the file after, and the edit.

#### Scenario: the recorded bytes

- **WHEN** a value is changed in the old Quiqr and saved
- **THEN** the file it wrote SHALL be kept as the expected result, byte for
  byte, including its final newline

#### Scenario: ChatCMS performs the same edit

- **WHEN** ChatCMS loads the recorded model and the file before, and sets the
  same field to the same value
- **THEN** what it writes SHALL equal the recorded file after, byte for byte,
  apart from differences the recording names

#### Scenario: the neighbours do not move

- **WHEN** the edited value sits among values of other kinds
- **THEN** every one of those SHALL be unchanged in both files, so that a
  passing case cannot hide a field that was rewritten on the way through

#### Scenario: recording is deliberate and offline afterwards

- **WHEN** the tests run
- **THEN** the case SHALL be read from what was committed, with nothing
  started and nothing fetched

### Requirement: Every difference is either absent or named

Where ChatCMS deliberately does not do what the old Quiqr does, the recording
SHALL name that difference and say why. A difference nothing names SHALL fail,
and so SHALL a named difference that the old Quiqr has stopped making.

#### Scenario: a difference nobody chose

- **WHEN** what ChatCMS writes differs from the recording in a way the
  recording does not name
- **THEN** the test SHALL fail, saying where the two part ways

#### Scenario: a difference that was chosen

- **WHEN** the recording names a difference and says why
- **THEN** that difference alone SHALL be allowed, and the rest of the file
  SHALL still match byte for byte

#### Scenario: a named difference that stopped happening

- **WHEN** the old Quiqr no longer makes a difference the recording names
- **THEN** the test SHALL fail, because the recording is describing an
  application that no longer behaves that way

#### Scenario: a difference with no reason

- **WHEN** a named difference carries no reason
- **THEN** the test SHALL fail, because a difference nobody can explain is
  not a decision
