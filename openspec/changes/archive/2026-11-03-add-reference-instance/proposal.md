# Proposal

Bean: `chatcms-37w8`
(`.beans/chatcms-37w8--setup-local-quiqr-web-install-for-e2testing.md`), under
milestone `chatcms-0fml`.

## Why

The project's backwards-compatibility promise is that a form in ChatCMS does
to a value what the old Quiqr does to it. Only one half of that is checked
against the real thing: `scripts/generate-goldens.sh` runs the two serialiser
libraries the old app uses, which settles how bytes are written. What a field
does to a value on the way in and out is checked against our reading of the
old app's source, which is a belief, not an observation.

`internal/model` decodes all 28 field types the old app draws, but
`web/src/fields/Fields.tsx` has 17 cases and a text-box fallback, so seven
types are drawn as something else today. Before any of that can be called
right or wrong, there has to be something to compare with that can be started,
pointed at a site, and driven.

This change builds that: a reference instance of the old Quiqr, set up by one
script, with the kitchen sink template mounted in it, and one field carried
all the way round the circle. One string field is edited in the old
application, what it wrote is recorded, and a test in the gate asserts that
ChatCMS writes the same bytes for the same edit. One case proves the circle
closes; the remaining 27 types are the next epic's work, and they cost a
recording each once the circle exists.

Acceptance test served: none directly. It serves the golden category of
`docs/TESTING.md`, which exists for acceptance test 3 of
`docs/BRIEFING-M01.md` (a form saves what the old Quiqr would have saved).

## What Changes

- **A script, `scripts/reference.sh`**, with `up`, `stop` and `fresh`, in the
  shape `scripts/dev.sh` already uses. It clones `quiqr/quiqr-desktop` at a
  pinned commit, builds and runs its standalone edition inside that
  checkout's own `nix develop`, starts it on a port of its own, makes a user
  whose password is known, and mounts one site.
- **A pinned site**: `quiqr/quiqr-template-kitchen-sink`, which declares 26
  of the 28 field types, including the seven ChatCMS draws as a text box.
- **One place that holds the versions**, printed on every start: the Quiqr
  commit, the template commit, node, and Hugo. A behaviour that changes
  because a version moved must be visible as a version having moved.
- **A section in `docs/TESTING.md`** saying what the reference instance is an
  oracle for, and why it is not in the gate.
- **One recorded case, end to end.** The kitchen sink's `simpleDataFields`
  single edits `dataFields.md`, whose front matter holds a `string` value
  surrounded by twenty values of other kinds. One edit to that string is made
  in the old application, and the file it writes is committed as
  `test/fixtures/reference/`, with the model, the file before, the file
  after, and the edit.
- **A test in the gate** loads that recording, performs the same edit through
  `internal/forms`, and asserts byte equality. It needs nothing running and
  no network, so it is an ordinary `go test`, and it fails the gate when
  ChatCMS and the old application disagree.
- **Not in `nix flake check`: the recording.** Making a recording needs the
  network and npm, so it is run by hand, like `scripts/generate-goldens.sh`.
  Checking a recording is in the gate, like the goldens it produced.

## Capabilities

### New Capabilities

- `acceptance/reference-instance`: what a reference instance of the old Quiqr
  owns, what it pins, what it reports, and what it must never become. It is
  the counterpart of `acceptance/harness`, which says the acceptance suite
  reaches nothing outside itself; this one exists precisely to reach outside,
  so the rules that keep that honest are worth writing down.

### Modified Capabilities

- `content/serialisation`: the capability already says the expected bytes are
  produced by running the reference rather than asserted. It means the two
  serialiser libraries. One requirement is added: at least one case is
  recorded from the whole application, through its own form, because the
  libraries only decide how a value is written down, not what the value has
  become by the time they see it.

## Impact

- **Added**: `scripts/reference.sh`, a pins file next to it, one recorded
  case under `test/fixtures/reference/`, one Go test that reads it, and a
  section in `docs/TESTING.md`.
- **Untouched**: the frontend, the database, the flake checks and every
  existing test. Nothing the server does changes. If the new test fails on
  the first recording, that is a finding about ChatCMS, and it is reported
  rather than worked around.
- **Seams**: none. The change touches no `gitsync`, `runner`, `policy` or
  `auth` boundary; it adds a development tool beside them.
- **Working tree**: a state directory (`.reference-state/`, ignored) holding
  the checkout, node modules, the old app's configuration and its site. Two
  gigabytes of node modules live there, which is why `fresh` exists and why
  `up` does not reinstall when the state is already good.
- **Network**: required on the first `up` and after a version moves. Recorded
  as a deliberate exception to the project's usual rule, in the same sentence
  as `scripts/generate-goldens.sh`.

## Assumptions

- **The standalone edition, not Electron.** The old app grew a standalone
  server (`npm run start:standalone`, `QUIQR_CONF_DIR`, `PORT`, `HOST`,
  a user CLI) that serves its API and its own frontend from one Express
  server. The Electron app needs a display and offers nothing here.
- **The fork to pin is `quiqr/quiqr-desktop`**, as the bean asks.
  `mipmip/quiqr-desktop` is ahead by exactly one commit, which adds libstdc++
  to `LD_LIBRARY_PATH` in that repository's own dev shell. The script uses
  upstream's dev shell and exports that one variable itself, so the fork is
  not needed and upstream stays the thing we track.
- **A site is mounted by writing a file.** The old app globs
  `<conf>/sites/*/config.json`, so mounting the kitchen sink is writing one
  JSON file. Nothing in this change drives a browser to set the instance up.
- **The kitchen sink is enough to start with.** It declares 26 of the 28
  types. `easymde` is the one drawable type it omits; `file` and `image`
  appear in the old app's documentation with no component behind them.
  Neither is a problem now: the next epic decides whether to extend the
  template or carry a second site.
- **One case, and it is a string.** The simplest type, so that a failure
  means the circle is wrong rather than that the field is hard. The point of
  this change is the circle, not coverage.
- **One instance at a time.** Unlike `scripts/dev.sh`, this script has no
  named instances. Nothing needs two references at once, and the state
  directory is large.
