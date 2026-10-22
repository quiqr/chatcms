# platform/server-runtime Specification

## Purpose
Defines how the ChatCMS server process starts, how it is configured, how it
reports whether it is healthy, and how it stops, so that operators and the
deployment module can rely on the same contract.

## Requirements

### Requirement: Configuration comes from the environment

The server SHALL read its configuration from environment variables and SHALL
refuse to start when a required value is missing. Every variable SHALL have a
documented default except those that cannot have one safely.

#### Scenario: A required value is missing

- **WHEN** the server starts without a database connection string
- **THEN** it exits non-zero without opening a listener
- **AND** it writes a message naming the missing variable

#### Scenario: Optional values fall back to defaults

- **WHEN** the server starts with only the database connection string set
- **THEN** it listens on `127.0.0.1:8080`
- **AND** it uses `git` and `hugo` from the configured binary paths

#### Scenario: A secret is read from a file

- **WHEN** a configuration value is supplied as a `_FILE` variable naming a
  readable file
- **THEN** the server uses the file contents as the value
- **AND** the value never appears in logs or in any API response

### Requirement: The server listens and serves a JSON API

The server SHALL expose an HTTP listener serving a JSON API under the `/api`
prefix. Responses SHALL use a consistent error shape carrying a machine
readable code and a human readable message.

#### Scenario: An unknown API path

- **WHEN** a request is made to an API path that does not exist
- **THEN** the response status is 404
- **AND** the body is JSON with an error code and message

#### Scenario: A malformed request body

- **WHEN** a request carries a body that is not valid JSON
- **THEN** the response status is 400
- **AND** the body is JSON with an error code and message
- **AND** the message does not include the raw body

#### Scenario: A handler panics

- **WHEN** a handler panics while serving a request
- **THEN** the response status is 500 with the generic error shape
- **AND** the process stays alive and serves the next request
- **AND** the panic and its stack are logged with the request id

### Requirement: Every request is traceable

The server SHALL assign a request id to each request, return it in a response
header, and include it in every log line emitted while serving that request.

#### Scenario: A request id is returned

- **WHEN** any request is served
- **THEN** the response carries a request id header
- **AND** the log lines for that request carry the same id

#### Scenario: A caller supplies a request id

- **WHEN** a request arrives with a request id header
- **THEN** the server reuses that id rather than generating a new one

### Requirement: Health is reported honestly

The server SHALL expose an unauthenticated health endpoint reporting whether
the database is reachable and which schema version is applied.

#### Scenario: The database is reachable

- **WHEN** the health endpoint is requested and the database answers
- **THEN** the response status is 200
- **AND** the body names the applied schema version and the server version

#### Scenario: The database is unreachable

- **WHEN** the health endpoint is requested and the database does not answer
- **THEN** the response status is 503
- **AND** the body says the database is unreachable
- **AND** the body does not include the connection string

### Requirement: Shutdown is graceful

On receiving an interrupt or termination signal the server SHALL stop
accepting new connections, allow in-flight requests a bounded time to finish,
and then exit zero.

#### Scenario: A signal arrives while a request is in flight

- **WHEN** the process receives a termination signal during a request
- **THEN** that request completes normally
- **AND** new connections are refused
- **AND** the process exits zero within the shutdown timeout

#### Scenario: A request outlives the shutdown timeout

- **WHEN** a request is still running when the shutdown timeout expires
- **THEN** the process exits non-zero
- **AND** it logs that shutdown was forced

### Requirement: A request that changes something proves it was meant

A request that changes anything SHALL present a token that only pages served
by this site can read, so that another site cannot make a reader's browser act
on their behalf.

#### Scenario: A change without proof

- **WHEN** a state-changing request arrives with a session but no matching
  token
- **THEN** it is refused with 403 and the JSON error shape
- **AND** nothing is changed

#### Scenario: A change with proof

- **WHEN** the request presents the token matching the one it was given
- **THEN** it proceeds

#### Scenario: Reading needs no proof

- **WHEN** a request only reads
- **THEN** it needs no such token

#### Scenario: Signing in needs no session but still needs proof

- **WHEN** a sign-in request arrives
- **THEN** it needs the token, because it changes something
- **AND** it does not need an existing session
