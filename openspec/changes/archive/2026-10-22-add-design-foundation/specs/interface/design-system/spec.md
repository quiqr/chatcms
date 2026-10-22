## Purpose

Defines what the interface is made of: where colour and type come from, which
components exist, and the rules that keep ten screens one system rather than
ten dialects.

## ADDED Requirements

### Requirement: Colour and type come from tokens

Every colour, radius and typeface in the interface SHALL come from the design
tokens. A component SHALL NOT write a colour of its own.

#### Scenario: A component uses a token

- **WHEN** a component needs a surface, a text colour or a border
- **THEN** it names a token rather than a colour value

#### Scenario: A colour written by hand fails the build

- **WHEN** a component source carries a colour value written out
- **THEN** the frontend check fails and names the file and the value

#### Scenario: The tokens are the ones from the hand-off

- **WHEN** the tokens are compared with the design hand-off
- **THEN** every token the hand-off defines is present under the same name

### Requirement: The typefaces are served by the application

The interface SHALL serve its typefaces itself rather than fetching them from
another party, so that no reader's browser contacts a third party and the
build needs no network.

#### Scenario: No third-party font request

- **WHEN** the built application is inspected
- **THEN** it contains no request to a font service

#### Scenario: The typefaces are the designed ones

- **WHEN** the interface renders text
- **THEN** the typefaces are the ones the design names, with a fallback stack
  behind each

### Requirement: The components the designs use exist

The components named in the design hand-off SHALL exist with the variants the
designs use, so a screen is assembled rather than invented.

#### Scenario: A button has the variants the designs use

- **WHEN** a screen needs a primary, ink, outline or ghost button
- **THEN** each exists and is styled as the designs show

#### Scenario: A status badge carries its meaning

- **WHEN** a status badge shows a changed, new, published, live or
  needs-attention state
- **THEN** each has its own soft background and dark text, and the ones the
  design gives a dot have one

#### Scenario: A component is reachable by keyboard

- **WHEN** a component that can be operated is rendered
- **THEN** it is a real control rather than a styled container
- **AND** an icon-only control carries a readable name

### Requirement: Touch targets are large enough

A control meant to be used by touch SHALL be at least forty-four pixels in its
smaller dimension; a secondary control on a pointer device may be smaller, and
not below thirty-six.

#### Scenario: A primary control

- **WHEN** a control is one a content creator uses on a phone
- **THEN** it is at least forty-four pixels high

#### Scenario: The rule is checked

- **WHEN** a control is defined with a height below the minimum for its kind
- **THEN** the frontend check fails and names it
