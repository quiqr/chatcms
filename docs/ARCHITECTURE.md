# Architecture

This is the shape the PoC is built into. It exists so every epic knows where
its code goes and which seams it must not cross. Package names are the ones the
coverage gate and the specs refer to, so keep them.

## The one rule

**Content lives in git. Everything else lives in Postgres.** The database never
holds page content, front matter or media bytes. Git never holds users, roles,
credentials or job state. When a feature seems to need both, it needs a
projection: a cached, rebuildable view in Postgres of something git owns.

## Go packages

| Package                | Responsibility                                                              |
| ---------------------- | --------------------------------------------------------------------------- |
| `internal/app`         | Configuration, wiring, start-up, migration run, HTTP server lifecycle         |
| `internal/httpapi`     | HTTP handlers, request and response types, session middleware                 |
| `internal/auth`        | Email and password login, password hashing, session issue and revoke          |
| `internal/policy`      | The single `Can(actor, action, resource)` function. Core package.             |
| `internal/tenancy`     | Organisations, projects, memberships, invitations                             |
| `internal/model`       | Reading a site's Quiqr model into a typed schema, and the navigation it implies. Core package. |
| `internal/forms`       | Mapping a document's values onto a model's fields and back. Core package.    |
| `internal/frontmatter` | Reading a content file, and writing front matter exactly as the reference does |
| `internal/content`     | Collections and singles: list, create, rename, delete                         |
| `internal/media`       | Upload, listing, thumbnails, alt text, usage lookup                           |
| `internal/gitsync`     | The repository engine: clone, commit, fetch, push, conflict detection. Core package. |
| `internal/secrets`     | Encryption at rest, and generating deploy keys. The thing a hardware key store would replace |
| `internal/onboarding`  | Bringing a customer's site in, its key, and where it publishes                |
| `internal/redact`      | Removing credential-shaped text before it is logged, stored or shown          |
| `internal/preview`     | Per-session Hugo preview servers, including idle shutdown                     |
| `internal/runner`      | Sandboxed build execution behind an interface                                 |
| `internal/sitebuild`   | Running the pinned Hugo over a site, through a runner                         |
| `internal/builds`      | What is remembered about every build, and the state it ended in               |
| `internal/workflows`   | Parsing and running the declarative YAML workflows                            |
| `internal/publish`     | The change summary, the publish pipeline, the publish record                  |
| `internal/history`     | Version listing in plain language, and rollback                               |
| `internal/jobs`        | Background job queue and worker, backed by Postgres. Knows nothing about sites |
| `internal/store`       | sqlc-generated query code plus thin helpers. Generated; do not hand-edit.     |
| `internal/version`     | The version string stamped by the build                                       |

## Seams that must stay thin

These are the places where milestone 1 deliberately picks the simple option and
must leave the door open. Each is an interface with exactly one implementation
today. Nothing outside the package may depend on the concrete type.

**`internal/gitsync.Repo`** wraps the git CLI. Every repository operation goes
through this interface so Jujutsu can replace git later. Nothing else in the
tree runs git, which a test over the sources enforces rather than leaving to
review. The git binary comes from `CHATCMS_GIT_BIN`, never from `PATH`.

Git runs with hooks disabled, no credential helper, no terminal prompt and no
system or user configuration. A customer's repository may contain a hook, and
a hook is code; running it on the server because a repository was imported is
not a thing this product does. A site that needs a step to run at a particular
moment declares it in `quiqr/workflows/`, which is the supported way and the
one that runs in the sandbox.

The interface speaks in sites and versions. No method takes a commit id, a
branch name or a refspec, so a caller can show history and roll back without
knowing what the identifier is. The cost is that a git-specific capability
that does not fit cannot be reached from outside; when one is genuinely
needed, it is added as a method in this vocabulary rather than by handing out
the repository path.

**`internal/runner.Runner`** executes a build in isolation. Milestone 1 ships a
bubblewrap implementation. Firecracker micro virtual machines replace it
later, so the interface speaks in terms of inputs, a command and collected
outputs, never in terms of local paths that only make sense in-process.

A task is inputs, a command, limits and the paths it wants collected. A result
is an exit code, a log, those outputs and which limit was reached. Nothing in
either names a process, a mount or a path on this machine.

The implementation needs `bwrap` and `prlimit` on the machine, both named by
settings and both in the NixOS module's path. If it cannot isolate, it refuses
to run anything, and the server says at start-up that builds are unavailable.
A build that quietly ran unsandboxed would be the worst available outcome: no
protection, and no sign that there is none.

What replacing it costs, and buys:

| | bubblewrap | A machine of its own |
| --- | --- | --- |
| Isolation from the host | Namespaces, so a kernel bug is a host bug | A kernel of its own |
| Memory limit | `prlimit` on the data segment, which is why a failure is read out of the log | Asked of the hypervisor, exactly |
| Start-up | A few milliseconds | Tens to hundreds of milliseconds |
| Where it runs | This machine | Anywhere |

The interface is already the second one's shape. What changes is the
implementation and the two tool settings; no caller moves.

**`internal/builds`** keeps the record: the project, the reason, the start and
finish, the state and the log. Builds left running when the server stopped are
marked interrupted at the next start, because a record that stays running
forever is worse than one that says it does not know.

**`internal/policy.Can`** is the only authorisation decision point in the
codebase. Handlers ask it; they never compare roles themselves, which a test
over the sources enforces by failing when any other package carries a role as
a value. The packages allowed to carry one are listed in that test, so adding
an exception is visible in a diff.

`Can` is pure: no database, no context, no error. The caller loads the
memberships; this package decides. That is what makes the whole role table
testable with the answers written out rather than worked out, which is how the
table from `docs/design/DESIGN.md` section 6 is checked.

Per-document permissions are out of scope for milestone 1, so `Can` takes a
resource argument it mostly ignores today. Keep the argument. Owning an
organisation carries the owner role on every project in it, so a project
cannot end up with nobody who can administer it.

**`internal/secrets`** seals everything that must not be readable in the
database, with AES-256-GCM under one master key read from the file
`CHATCMS_SECRET_KEY_FILE` names. Deploy keys are Ed25519, generated in process
rather than by running `ssh-keygen`, so no private key is ever written to a
temporary file on the way.

Git reaches a remote by running `ssh -i <file>`, so a key has to be on disk
while it is used. One directory under the state directory holds them, made
owner-only when the server starts and removed when it stops. That is the cost
of not writing an agent, and it is named here rather than discovered later.

**`internal/auth`** issues sessions. OIDC is out of scope, so the identity
source is an interface with one implementation: `Identity.Authenticate` takes
a `Claim`, which a password satisfies today and an assertion from a provider
could satisfy later. Sessions do not assume a password ever existed, which is
what an invited person looks like before they accept.

A session is 32 random bytes in a cookie and a SHA-256 of them in the
database, so a copy of the database grants nobody a session. Passwords are
argon2id with the parameters written into the stored value, so raising them
later leaves old hashes verifiable. A sign-in always issues a new session and
revokes whatever the request arrived with.

## Request path

```
HTTP request
  -> internal/httpapi    session lookup, request decoding
  -> internal/policy     Can(actor, action, resource)
  -> feature package     content, media, forms, publish, ...
  -> internal/gitsync    working copy mutation, if any
  -> internal/jobs       queue the sync, build or publish work
```

A write to content is: validate against the model, write the file in the
working copy, auto-commit through `gitsync`, queue a sync job. The user sees
"saved" when the working copy write succeeds, not when the push lands. The sync
indicator reports the rest.

## Frontend

`web/` is a Vite app. It talks to the Go server over a JSON API under `/api`.
The server serves the built assets from an embedded filesystem in production,
so a deployment is one binary plus Postgres plus git plus Hugo.

Form rendering is driven by the schema the server derives from the Quiqr model,
not by hand-written forms per site. A new field type means a new renderer in
`web/src/fields/` and a new case in `internal/model`, nothing else.

## What the database holds

Users, organisations, memberships, roles, project records, encrypted deploy
keys, sessions, invitations, job state, publish records and cached projections
of git state (such as the unpublished change count). Everything in the last
category must be rebuildable from git alone.
