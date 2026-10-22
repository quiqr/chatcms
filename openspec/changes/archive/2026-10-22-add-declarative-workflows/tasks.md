# Tasks

## 1. The format

- [x] 1.1 Types: trigger, action, step, workflow
- [x] 1.2 Parse one file, strictly: unknown keys are refused
- [x] 1.3 Validate: known trigger, known action, `run` only on the check,
      publishing last and once, building once
- [x] 1.4 Errors that name the file, the step and the key
- [x] 1.5 Read a directory: ignore what is not YAML, refuse two for one trigger

## 2. Defaults

- [x] 2.1 Work the default out from the project's publishing settings
- [x] 2.2 A declared workflow with no steps means nothing, not the default

## 3. Running

- [x] 3.1 The engine, its dependencies and the result types
- [x] 3.2 Steps in order, stopping at the first failure
- [x] 3.3 `check` through the runner: site files, fixed toolbox, no network, limits
- [x] 3.4 `build` through `internal/sitebuild`, output carried to the next step
- [x] 3.5 `publish` through the publisher seam, never with a command
- [x] 3.6 Release what a build produced whatever happened

## 4. What comes back

- [x] 4.1 A sentence for the editor, with no git word in it
- [x] 4.2 The raw output kept apart and bounded
- [x] 4.3 A test that the editor-facing sentences carry no git word
- [x] 4.4 Step by step: name, action, state, duration

## 5. Shipping

- [x] 5.1 Settings for the toolbox and the check limits
- [x] 5.2 The toolbox in the dev shell, the checks and the NixOS module
- [x] 5.3 A fixture site with a workflow, and one with a failing check
- [x] 5.4 `docs/ARCHITECTURE.md`: the workflow format and what is sandboxed
- [x] 5.5 The gate green, coverage held
