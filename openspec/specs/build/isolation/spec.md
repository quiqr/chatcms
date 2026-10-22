# build/isolation Specification

## Purpose
Defines what a build can reach, what happens when it exceeds a limit, and what
of its work comes back, so that running somebody else's build code is a bounded
risk rather than a trusted act.

## Requirements

### Requirement: A build sees only what it was given

A build SHALL run against a tree of input files and a named set of tools, and
SHALL NOT be able to read any other part of the machine: no other project's
working copy, no state directory, no deploy key, no home directory, no
environment of the server process.

#### Scenario: another project's working copy is not readable

- **WHEN** a build runs on one project while another project's working copy
  exists on the same machine
- **THEN** the other project's files SHALL NOT be readable from inside the
  build
- **AND** this SHALL hold whether the build reads the path it was given or an
  absolute path it guessed

#### Scenario: the deploy keys are not readable

- **WHEN** a build tries to read the directory where decrypted deploy keys are
  written
- **THEN** it SHALL fail to read it

#### Scenario: the server's environment is not inherited

- **WHEN** a build runs
- **THEN** it SHALL see only the environment entries the task named
- **AND** SHALL NOT see any variable set for the server process

### Requirement: A build cannot reach the network unless it is asked for

A build SHALL have no network access by default. A task MAY ask for network
access, and a runner MAY refuse to grant it.

#### Scenario: a build with no request cannot reach the network

- **WHEN** a build that did not ask for network access tries to open a
  connection
- **THEN** the attempt SHALL fail

#### Scenario: asking is recorded rather than assumed

- **WHEN** a task asks for network access
- **THEN** that SHALL be a stated property of the task rather than a property
  of how the runner happens to be configured

### Requirement: A build cannot write outside its own workspace

A build SHALL be able to write to its own working directory and to a temporary
area, and SHALL NOT be able to write anywhere else.

#### Scenario: writing outside the workspace fails

- **WHEN** a build tries to write to the state directory, to the store, or to
  any path above its working directory
- **THEN** the write SHALL fail
- **AND** the build SHALL still be able to write its own output

### Requirement: A build ends when it exceeds a limit

Every build SHALL carry a time limit and a memory limit. Exceeding either SHALL
end the build and SHALL be reported as a limit having been reached rather than
as an ordinary failure.

#### Scenario: a build that runs too long is ended

- **WHEN** a build runs for longer than its time limit
- **THEN** it SHALL be ended
- **AND** the result SHALL say that the time limit was reached
- **AND** the output produced before it was ended SHALL be kept

#### Scenario: a build that asks for too much memory is ended

- **WHEN** a build tries to use more memory than its limit
- **THEN** it SHALL not succeed
- **AND** the result SHALL carry a log

#### Scenario: a build that says it could not allocate is reported as such

- **WHEN** a build fails and its output says plainly that it could not
  allocate memory
- **THEN** the result SHALL report the memory limit as what was reached
- **AND** where the build says nothing recognisable, the result SHALL report
  an ordinary failure rather than guessing

#### Scenario: a cancelled build leaves nothing running

- **WHEN** the caller cancels a build
- **THEN** no process from that build SHALL still be running when the call
  returns
- **AND** the workspace it used SHALL be removed

### Requirement: What a build produced is collected by name

A task SHALL name the paths it wants collected, relative to the working
directory. Only those paths SHALL be readable by the caller afterwards.

#### Scenario: a named output is returned

- **WHEN** a build writes a tree and the task named that tree
- **THEN** every file in it SHALL be readable from the result

#### Scenario: an unnamed path is not returned

- **WHEN** a build writes a file the task did not name
- **THEN** that file SHALL NOT appear in the result

#### Scenario: a missing output is an error rather than an empty result

- **WHEN** a build succeeds but did not produce a path the task named
- **THEN** the call SHALL report that the output is missing
- **AND** SHALL NOT report success with nothing in it

#### Scenario: an output path cannot escape the workspace

- **WHEN** a task names an output path that climbs above the working directory
  or is absolute
- **THEN** the task SHALL be refused before anything runs

### Requirement: A runner that cannot isolate refuses to run

A runner SHALL check that it can isolate a build before it runs one, and SHALL
refuse rather than run a build without isolation.

#### Scenario: no sandbox means no build

- **WHEN** the means of isolation is absent or does not work on this machine
- **THEN** every build SHALL be refused with an error saying so
- **AND** no build command SHALL be run

#### Scenario: the check happens once rather than per build

- **WHEN** many builds run
- **THEN** the cost of checking SHALL not be paid for each of them

### Requirement: The runner speaks in inputs, a command and outputs

The interface a caller uses SHALL describe a build as inputs, a command, limits
and wanted outputs. It SHALL NOT mention processes, sandboxes, paths on the
host, or any other property of how isolation happens to be achieved.

#### Scenario: a second implementation needs no caller change

- **WHEN** a different means of isolation replaces the first
- **THEN** no caller SHALL have to change
- **AND** the interface SHALL be satisfiable by an implementation that runs the
  build on another machine

### Requirement: A site is built with the pinned Hugo

Building a site SHALL run the pinned Hugo through the runner, with the site as
input and the generated site as the collected output.

#### Scenario: the build output matches Hugo run directly

- **WHEN** a site is built through the runner
- **THEN** the generated files SHALL be the same as running the pinned Hugo on
  the same site directly

#### Scenario: a site that fails to build reports why

- **WHEN** a site has an error that makes Hugo fail
- **THEN** the build SHALL be reported as failed
- **AND** the log SHALL contain what Hugo said about it
