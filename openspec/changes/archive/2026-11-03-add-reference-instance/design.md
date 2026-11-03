# Design

## Context

See `proposal.md` for why. What matters here is what the old Quiqr now is: an
npm workspace with two editions. The Electron one needs a display. The
standalone one (`npm run start:standalone`) builds three workspaces and
starts an Express server that serves both its API and its own frontend, taking
`QUIQR_CONF_DIR`, `PORT` and `HOST` from the environment.

Two facts decide most of this design:

- A site is registered by writing a file.
  `packages/backend/src/services/configuration/configuration-data-provider.ts`
  globs `<conf>/sites/*/config.json` and reads
  `{key, name, source: {type: "folder", path}}`.
- A user is a row in `<conf>/users.json` with a bcrypt hash, made by
  `quiqr-admin add <email>`, which reads the password from stdin.

So the whole set-up is files and environment variables. Nothing needs a
browser.

## Goals / Non-Goals

**Goals**

- One command, one state directory, one file of pins.
- An answer read off the reference is read off the filesystem, not off a
  screen.
- Two people running this on different machines see the same application.

**Non-Goals**

- Packaging the old Quiqr with Nix. A derivation would be the right answer if
  the recording step ever belonged in the gate; it does not.
- Running more than one reference at a time.
- Coverage. One field type is carried round the circle. The other 27 are the
  next epic, and this change exists to make each of them cheap.

## Decisions

### The standalone edition, not Electron

The behaviour in question lives in the backend and the form components, which
both editions share. Electron adds a display requirement and `xvfb` to the
set-up and gives nothing back.

*Alternative considered*: the repository's `docker-compose.yml`, which starts
the same edition. Rejected because it needs a daemon the project does not
otherwise use, and because the sites directory then lives inside a volume,
which is the opposite of what we want: the files the old app writes are the
observation.

### A git clone into a state directory, not a flake input

The checkout is fetched by the script into `.reference-state/`, not declared
as a flake input.

- A flake input is evaluated by every `nix` command in the project, so the
  old application would become a dependency of work that has nothing to do
  with it.
- Flake inputs from GitHub need a token here, and that token expires. A
  development tool that stops working because an unrelated credential lapsed
  is worse than one that needs the network when it is deliberately run.
- The install is gigabytes of node modules. That is state, not a build
  product worth putting in the store.

*Alternative considered*: a derivation that builds the standalone edition.
Better in every way except the three above, and worth revisiting if a
comparison ever has to run inside `nix flake check`. Writing it down so the
option stays open.

### The mounted site is a git clone, and `git diff` is the instrument

The kitchen sink is cloned at a pinned commit into the state directory and
mounted from there. After an edit through the old app, `git -C <site> diff`
is exactly what the application did to the files, with no interpretation in
between. A site copied out of a tarball would need a second copy kept beside
it to compare with.

### The user is made by the CLI, and the forced change is cleared

`quiqr-admin add` creates a user with `mustChangePassword: true`, which would
stop any first use until somebody typed a new password. The script pipes the
password into the CLI and then sets that flag to false in `users.json`. A
reference that cannot be set up by one command is not set up by one command.

*Alternative considered*: letting the first run create its default admin and
parsing the credentials out of the log. Rejected: the format of a log line is
not a contract, and the forced change would still be there.

### The old app's own flake, plus the one line it is missing

Every build and run step happens inside `nix develop` on the pinned checkout,
as the bean asks. That dev shell is not a convenience: it sets `EMBGIT_PATH`
and `HUGO_PATH` and builds embgit from `pkg-embgit.nix`. Embgit is how the old
application talks to git, so supplying node ourselves would mean building
embgit ourselves, which is a second place that knows how the old app reaches
a repository.

Upstream's dev shell lacks the one export that `mipmip/quiqr-desktop` adds:
libstdc++ on `LD_LIBRARY_PATH`, without which the prebuilt native modules
(sharp, libvips) fail to load. The script exports it inside the shell. That is
why upstream can be preferred, as the bean asks, without the fork: the
difference between them is a single line we can carry ourselves, and carrying
it keeps us tracking the repository the code actually lands in.

What the shell gave (node, Hugo, embgit) is printed at start-up with the
pins.

### Ports away from the defaults

The reference takes a port of its own rather than 5150, and gives the old
app's preview a port other than 13131, so that a developer running their own
Quiqr does not have to stop it first.

### One file holds the pins

`scripts/reference-pins.env`: the Quiqr commit, the template commit, and the
ports. Moving a pin is one edit, visible in a diff, which is what makes a
changed answer traceable to a changed version.

### The circle, and where each half runs

```
   by hand, needs the network          in the gate, offline
  ┌──────────────────────────┐       ┌────────────────────────┐
  │ reference.sh up          │       │ go test ./internal/... │
  │   old Quiqr + kitchen    │       │                        │
  │   sink, pinned           │       │  model + before        │
  │                          │       │       │                │
  │ edit "string" in         │       │       ▼                │
  │ simpleDataFields, save   │       │  set string = V        │
  │       │                  │       │       │                │
  │       ▼                  │       │       ▼                │
  │ git diff → record  ──────┼──────▶│  bytes == after ?      │
  └──────────────────────────┘commit └────────────────────────┘
```

The left half is a person, once. The right half is the gate, every time.

### The case is a string in `dataFields.md`

The kitchen sink's `simpleDataFields` single edits `dataFields.md`, whose
front matter holds `string: Single line of text` among twenty values of other
kinds: chips, a number, a multiline string, a uniq, a readonly, a boolean, a
leaf array, selects, a date, a colour, two pickers.

That is what makes one case worth having. The assertion is byte equality over
the whole file, so the case fails if the edited value is written differently
*or* if any neighbour is rewritten on the way through. A string field alone
in a file would prove far less.

### Recording is files, not a format we invent

A recording is a directory: the model as it was, the file before, the file
after, and one small file saying which field was set to what, plus the
versions that produced it. No parser of ours stands between the old
application's output and the assertion.

*Alternative considered*: driving the old app's HTTP API and recording its
responses. Rejected: the API is below the form components, and the form
components are the part in question.

## Risks / Trade-offs

- **The kitchen sink is from December 2024 and the old app has moved.**
  → The first task is to bring it up and write down what the application says
  about the model, before anything else is built on it.
- **A later Quiqr commit changes the first-run or user flow.** → The pin
  protects the script; moving the pin is a deliberate act, and the script
  fails loudly rather than guessing if `users.json` is not what it expects.
- **npm install needs the network and several minutes.** → It happens once;
  `up` does nothing when the state is good; `fresh` is the way to redo it.
- **Two gigabytes in the working tree.** → One ignored directory, removable
  with one command, named in the documentation.
- **`nix develop` on the old checkout fetches nixpkgs-unstable**, so the first
  `up` needs whatever credentials this machine's Nix uses for GitHub. That is
  the same credential the project's own flake needs, so a failure here is the
  same failure, fixed in the same place, and the script says so rather than
  reporting a build error.
- **The first recording may fail against ChatCMS.** That is the change
  working, not the change broken. The finding is reported and a decision
  taken; the test is not loosened to make it pass.
- **One case can lull.** Twenty-seven types stay unchecked, and a green gate
  will not say so. The documentation says which type is covered, and the
  milestone exists to finish the job.
- **The reference drifts from what a customer runs.** → It is pinned to a
  commit of the organisation's repository, and prints it. Which commit a
  customer runs is a separate question, and a better one to ask once there is
  something to compare with.

## Migration Plan

None. Nothing existing changes: a script, a pins file and a documentation
section are added. Removing them removes the feature.

## Open Questions

- Whether the kitchen sink is extended to cover `easymde` (and `file` and
  `image`, which the old app documents with no component behind them), or a
  second site is carried for them. The next epic has the evidence to decide;
  this one records the gap.
- Whether the old app's own API is worth driving directly rather than its
  frontend, once there is a comparison to make. Both read and write the same
  files, so the choice does not change this design.
