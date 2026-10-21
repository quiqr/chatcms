# content/serialisation Specification

## Purpose
Defines the bytes ChatCMS writes, and the guarantee that they are the bytes
the Quiqr desktop app would have written, so that adopting ChatCMS does not
rewrite a customer's repository.

## Requirements

### Requirement: A document that is read and written back is unchanged

Reading a content file the reference implementation wrote and writing it back
with no edit SHALL produce exactly the same bytes.

#### Scenario: A round trip with no edit

- **WHEN** a content file produced by the reference implementation is read and
  written back unchanged
- **THEN** the bytes are identical, including the final newline

#### Scenario: A body with its own delimiters

- **WHEN** a body contains a line that looks like a front matter delimiter
- **THEN** a round trip leaves it where it was

### Requirement: A format that cannot be written faithfully is refused

Front matter in a format the system cannot write the way the reference writes
it SHALL be readable, and writing it SHALL fail with a message naming the
format, rather than producing a file that differs from what the reference
would have produced.

#### Scenario: Reading a format that cannot be written

- **WHEN** a content file uses front matter in such a format
- **THEN** its values and body are readable

#### Scenario: Writing it is refused

- **WHEN** such a document is written
- **THEN** writing fails with a message naming the format
- **AND** no file is produced

#### Scenario: The format every site uses is writable

- **WHEN** a document whose front matter is YAML is written
- **THEN** it succeeds

### Requirement: Written values match the reference implementation

Values written into front matter SHALL be spelled as the reference
implementation spells them.

#### Scenario: A scalar that needs no quoting

- **WHEN** a string is written that cannot be mistaken for another type
- **THEN** it is written without quotes

#### Scenario: A scalar that would otherwise change meaning

- **WHEN** a string is written that reads as a number, as true or false, or as
  a word with another meaning in the format
- **THEN** it is quoted so that reading it back yields the same string

#### Scenario: A string with more than one line

- **WHEN** a multi-line string is written
- **THEN** it is written as a block rather than with escape sequences

#### Scenario: A string longer than the line width

- **WHEN** a single-line string is longer than the reference implementation's
  line width
- **THEN** it is wrapped the way the reference implementation wraps it
- **AND** reading it back yields the original string

#### Scenario: Nothing, and empty things

- **WHEN** a value is absent, an empty string, an empty list or an empty
  object
- **THEN** each is written the way the reference implementation writes it, and
  the four are distinguishable from one another

#### Scenario: Lists and nested objects

- **WHEN** a list of values, a list of objects, or a nested object is written
- **THEN** the indentation and layout match the reference implementation

### Requirement: An edit changes only what was edited

Writing a document after changing one value SHALL leave every other value
spelled as it was.

#### Scenario: One field changed

- **WHEN** one value is changed and the document is written
- **THEN** the only difference from the original bytes is that value

#### Scenario: A field added

- **WHEN** a value is added for a field that had none
- **THEN** it appears in the order the model declares it among the values that
  are there
- **AND** no other value moves

#### Scenario: A field cleared

- **WHEN** a value is cleared
- **THEN** it is written the way the reference implementation writes an absent
  value rather than being silently dropped

### Requirement: A shape that cannot be reproduced is refused

Writing a value whose spelling has not been proven to match the reference
implementation SHALL fail, rather than writing something that differs.

#### Scenario: An unsupported shape

- **WHEN** a value is written whose kind the emitter has not been proven
  against
- **THEN** writing fails with an error naming the field and the kind
- **AND** nothing is written

#### Scenario: The refusal is specific

- **WHEN** such a failure happens
- **THEN** the message says which value and what kind it was, so it can be
  added deliberately rather than guessed at

### Requirement: The guarantee is checked against the reference, not asserted

The expected bytes SHALL be produced by running the reference implementation
itself and recorded, so that the test cannot agree with a mistaken belief
about what the reference does.

#### Scenario: Goldens come from the reference implementation

- **WHEN** the expected output for a case is needed
- **THEN** it was produced by running the reference implementation over that
  case and committed

#### Scenario: Regenerating is deliberate

- **WHEN** the tests run normally
- **THEN** no golden is regenerated
- **AND** a difference fails the gate
