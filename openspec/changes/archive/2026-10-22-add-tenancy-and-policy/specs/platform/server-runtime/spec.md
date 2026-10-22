## ADDED Requirements

### Requirement: A request the actor may not make is refused

An endpoint SHALL refuse a request whose actor may not make it, and SHALL NOT
reveal through the refusal whether the thing acted on exists.

#### Scenario: Someone who may not

- **WHEN** a signed-in person makes a request they have no permission for
- **THEN** the response is 403 with the JSON error shape
- **AND** nothing is changed

#### Scenario: Something that is not theirs

- **WHEN** a signed-in person names something belonging to an organisation
  they are not in
- **THEN** the response is the same as for something that does not exist

#### Scenario: Nobody at all

- **WHEN** an unauthenticated request reaches an endpoint that needs a person
- **THEN** the response is 401 rather than 403
