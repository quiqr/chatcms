## Context

`internal/model` reads a site from an `fs.FS` and never asks where it came
from. This change supplies that `fs.FS` from a real git working copy, and adds
everything needed to keep it level with a remote. See `proposal.md` for why it
comes before anything that writes content.

Two constraints shape the design. `AGENTS.md` locks the git CLI as the
implementation, not a Go library. And `docs/ARCHITECTURE.md` says the
repository layer must stay thin enough that Jujutsu can replace it, which
means the interface may not leak git's vocabulary or its failure modes.

## Goals / Non-Goals

**Goals:**

- Never lose an editor's work, including when the process dies mid-exchange.
- Keep a merge marker out of every file an editor can open.
- Make status a question answered by looking, not by remembering.
- Keep the interface narrow enough that a second implementation is plausible.

**Non-Goals:**

- Credentials. The onboarding epic owns deploy keys; this change takes a key
  path and uses it.
- Projects. There is no project table yet; a site is an id and a directory.
- Per-field conflict resolution. This change reports a conflicted file with
  its three versions; the form epic turns that into a question about a field.
- Two server instances writing one site. Named as a limit below.

## Decisions

### The interface speaks in sites and versions

```go
type Repo interface {
    FS() fs.FS
    Save(ctx context.Context, save Save) (Version, error)
    History(ctx context.Context, limit int) ([]Version, error)
    Status(ctx context.Context) (Status, error)
    Sync(ctx context.Context) (SyncResult, error)
    Publish(ctx context.Context) (PublishResult, error)
    Resolve(ctx context.Context, resolutions map[string][]byte) (SyncResult, error)
    Abandon(ctx context.Context) error
}
```

No method takes a commit id, a branch name or a refspec. `Version` carries an
opaque `ID`, a time, an author, a message and the paths it changed. A caller
can show history and roll back without knowing what the id is.

The cost of this is real and worth naming: a git-specific capability that does
not fit, such as cherry-picking or reflog archaeology, cannot be reached from
outside. That is the point. When one is genuinely needed, it is added as a
method in this vocabulary rather than by handing out the repository path.

Alternative considered: exposing a thin `Git` type and letting callers compose
operations. Rejected because every caller would then encode its own idea of
what a sync is, and the second implementation would be impossible.

### Every git call goes through one runner, and a test enforces it

A single `runner` type builds the argument list, sets the environment and
executes. It always passes `-c` settings that make git deterministic and
non-interactive:

| Setting                      | Why                                                     |
| ---------------------------- | -------------------------------------------------------- |
| `core.hooksPath=/dev/null`   | a hook in a customer repository must not run on our server |
| `credential.helper=`         | never consult a helper, never prompt                     |
| `GIT_TERMINAL_PROMPT=0`      | fail instead of waiting for input that will never come   |
| `GIT_CONFIG_NOSYSTEM=1`      | the server's global git config must not change behaviour |
| `HOME` set to the state dir  | same reason                                              |
| `GIT_SSH_COMMAND`            | use the site's key, and no agent or known-hosts surprise |

The enforcement test walks the package sources and fails when any package
other than `internal/gitsync` mentions executing git. A rule like this is
worth nothing if it is only written down.

### Rebase, and what it costs

Local versions are replayed onto the remote. History stays a list an editor
can read, which `docs/design/DESIGN.md` section 5.9 needs. The cost is that a
conflict surfaces once per replayed version rather than once in total, so an
editor could in principle be asked twice about the same file.

Mitigation: the engine replays the whole local run with the conflict stopping
at the first version, and the resolution applies to the rest of the replay
through git's own rerere-like reuse being switched off and the resolution
being recorded by the engine itself, so a second occurrence of the same path
with the same two sides is resolved the same way without asking again. If the
sides differ, asking again is correct.

Alternative considered: merging, which asks once. Rejected because every sync
with concurrent edits would add a merge version, and the history screen would
show versions an editor did not make and cannot explain.

### A conflict is extracted, never left on disk

When a replay stops with a conflict, the engine reads all three versions of
each conflicted path out of git's index stages and then **resets the working
tree**, so the files on disk are never the ones with markers in them. The
three versions live in memory, in the `Conflict` value handed to the caller.

This is the decision that keeps the promise that no editor ever sees a marker.
It costs memory proportional to the conflicted files, which for Markdown
content is nothing worth optimising.

Resolution writes the chosen content, stages it, and continues the replay.
Abandoning aborts the replay, which git restores exactly.

### Interrupted exchanges are recoverable because git already makes them so

A replay in progress is visible in the repository itself. On open, the engine
asks git whether one is in progress and reports it. The caller can abandon,
which restores the pre-exchange state, or resolve and continue. No state of
our own is written, so there is no second source of truth to get out of step
with the first.

### Status is three questions, all asked of git

| Question                    | How it is answered                         |
| --------------------------- | ------------------------------------------- |
| Is there unsaved work?      | the working tree against the last version   |
| How many are unpublished?   | counting versions the remote lacks          |
| Did the remote move?        | comparing against the last fetched remote state |

Nothing is cached. The epic asks for a projection into PostgreSQL for the
unpublished count; that is an optimisation of this computation and lands with
projects, because a projection needs something to project onto.

### One writer per site, in this process

A per-site mutex, held across the whole of a save or an exchange. Two servers
writing one working copy is not possible anyway, because they would share a
directory; the deployment model is one server. This is written down rather
than enforced because enforcing it means a lock the deployment epic would have
to own.

### Jobs are a table, a claim and a worker

`internal/jobs` uses the `jobs` table the foundation added. Claiming is the
`FOR UPDATE SKIP LOCKED` query already generated by sqlc, which makes
exclusivity the database's problem rather than ours.

Deduplication is by kind and subject, enforced by a unique index over queued
rows rather than by a read-then-write in Go, so two requests arriving at once
cannot both pass a check and both insert. Queueing a sync for a site that
already has one queued is a no-op; queueing one for a site whose sync is
*running* enqueues exactly one more, because a running row is not in the
index. That is the difference between an editor's change being synced and
sitting unsynced until the next schedule.

Retries use exponential backoff with a cap and a maximum attempt count, after
which the job is recorded failed with its last reason. The reason is scrubbed
of anything resembling a credential before it is stored, because a git error
message can contain a URL with a token in it.

## Risks / Trade-offs

- **A rebase can ask about one file twice.** Mitigated by remembering
  resolutions within one exchange, as described above. A genuinely different
  conflict on the same path is a different question and is asked.
- **Reading conflict stages then resetting is two steps.** If the process dies
  between them the working tree still holds markers. Mitigated by the recovery
  path: opening a working copy with an exchange in progress reports it, and
  the content filesystem refuses to serve a file that is in conflict rather
  than serving one with markers.
- **Hooks are disabled in customer repositories.** A site relying on a git
  hook will behave differently than it does on a developer's machine.
  Mitigated by the workflows epic, which is the supported way to run a step,
  and by saying so in the documentation.
- **A very large repository makes the first clone slow.** No mitigation in
  this change beyond doing it as a job, so nothing blocks on it. Shallow
  clones are deliberately avoided: history is a product feature here.
- **The git version is pinned by Nix** and the engine depends on behaviour
  that varies between versions, such as porcelain output. Mitigated by using
  the stable machine-readable forms, and by the gate running against the
  pinned binary.

## Migration Plan

Nothing is deployed. The `jobs` table from the foundation change gains a
subject column and a unique index over queued rows. The migration is
reversible and is proven so by the migration test.

## Open Questions

- Whether the sync schedule should be per site or global. It is global and
  configurable here, because there are no projects to configure it on yet, and
  moving it to a per-project setting is additive.
