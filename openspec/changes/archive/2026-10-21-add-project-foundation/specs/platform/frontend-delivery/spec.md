## Purpose

Defines how the browser application reaches the user: built reproducibly,
shipped inside the server binary, and served so that a deployment is one
binary rather than a binary plus a directory of assets.

## ADDED Requirements

### Requirement: A deployment is one binary

The server binary SHALL contain the built frontend assets and SHALL serve them
without reading any asset from disk.

#### Scenario: Serving from an empty directory

- **WHEN** the binary is run from a directory containing nothing
- **THEN** requesting the application root returns the application document

#### Scenario: A client-side route

- **WHEN** a browser requests a path that is not an API path and not an asset
- **THEN** the server returns the application document with status 200
- **AND** the browser-side router decides what to render

#### Scenario: An unknown API path is not the application

- **WHEN** a request is made to an unknown path under the API prefix
- **THEN** the server returns the JSON 404 error shape, not the application
  document

### Requirement: Assets are cacheable and versioned

Built asset filenames SHALL carry a content hash, and the server SHALL serve
them with a long-lived cache policy while serving the application document
with a no-cache policy.

#### Scenario: An asset is requested

- **WHEN** a hashed asset file is requested
- **THEN** the response carries a long-lived immutable cache header

#### Scenario: The application document is requested

- **WHEN** the application document is requested
- **THEN** the response forbids caching
- **AND** a redeployed binary is picked up on the next page load

### Requirement: The frontend is built reproducibly

The frontend SHALL be built by the project's build system from a pinned
dependency tree, and the build SHALL be part of the gate.

#### Scenario: The gate builds the frontend

- **WHEN** the gate runs
- **THEN** the frontend is built from its pinned dependency tree
- **AND** a frontend type error or test failure fails the gate

#### Scenario: A dependency changes without its pin

- **WHEN** the dependency manifest changes but its pinned hash does not
- **THEN** the build fails rather than silently fetching something else

### Requirement: Development serves the same API

In development the frontend SHALL be served by its own dev server and SHALL
reach the same API paths as in production, without code in the application
knowing which mode it runs in.

#### Scenario: A request from the dev server

- **WHEN** the application running under the dev server calls an API path
- **THEN** the request reaches the running server process
- **AND** the application uses the same path it uses in production
