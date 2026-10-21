# platform/background-jobs Specification

## Purpose
Defines how work that must not block a request, such as talking to a remote
repository, is queued, picked up, retried and observed.

## Requirements

### Requirement: Work is queued, not awaited

Work that may be slow SHALL be queued and acknowledged immediately, so that no
request waits for a remote system.

#### Scenario: Queueing returns at once

- **WHEN** a job is queued
- **THEN** the call returns with the job's identifier without running it

#### Scenario: A queued job survives a restart

- **WHEN** a job is queued and the process stops before running it
- **AND** the process starts again
- **THEN** the job is still queued and is picked up

### Requirement: A job is claimed by exactly one worker

Claiming a job SHALL be exclusive, so that two workers never run the same job
at the same time.

#### Scenario: Two workers, one job

- **WHEN** two workers claim at the same moment and one job is queued
- **THEN** exactly one of them gets it
- **AND** the other is told there is nothing to run

#### Scenario: Nothing to claim

- **WHEN** a worker claims and the queue is empty
- **THEN** it is told so rather than failing

### Requirement: A failed job is retried with growing delay

A job that fails SHALL be retried after a delay that grows with each attempt,
up to a limit, after which it SHALL be recorded as failed with the reason.

#### Scenario: A transient failure

- **WHEN** a job fails once and is retried
- **THEN** it does not run again immediately
- **AND** it runs again after its delay has passed

#### Scenario: Repeated failure

- **WHEN** a job fails as many times as its limit allows
- **THEN** it is recorded as failed
- **AND** the last reason is readable

#### Scenario: The reason carries no secret

- **WHEN** a job fails for any reason
- **THEN** the recorded reason contains no credential and no connection string

### Requirement: A job that is already queued is not queued twice

Queueing work that is already waiting for the same subject SHALL not create a
second job.

#### Scenario: Two syncs asked for at once

- **WHEN** a sync is queued for a site that already has one queued
- **THEN** the queue still holds one job for that site

#### Scenario: A sync asked for while one is running

- **WHEN** a sync is queued for a site whose sync is already running
- **THEN** one further job is queued, so the request is not lost

### Requirement: Shutdown drains rather than abandons

On shutdown the worker SHALL stop claiming new jobs, let the job it is running
finish within the shutdown budget, and return unfinished work to the queue.

#### Scenario: A job in flight at shutdown

- **WHEN** the process is asked to stop while a job is running
- **THEN** that job runs to completion if it fits in the shutdown budget
- **AND** no new job is claimed

#### Scenario: A job that outlives the budget

- **WHEN** a running job does not finish within the shutdown budget
- **THEN** the job is returned to the queue rather than recorded as done
- **AND** it is picked up after the next start
