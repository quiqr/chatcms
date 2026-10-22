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
| `internal/secrets`     | Encryption at rest for per-project SSH deploy keys                            |
| `internal/redact`      | Removing credential-shaped text before it is logged, stored or shown          |
| `internal/preview`     | Per-session Hugo preview servers, including idle shutdown                     |
| `internal/runner`      | Sandboxed build execution behind an interface                                 |
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
Nix sandbox implementation. Firecracker micro VMs replace it later, so the
interface speaks in terms of inputs, a command and collected outputs, never in
terms of local paths that only make sense in-process.

**`internal/policy.Can`** is the only authorisation decision point in the
codebase. Handlers ask it; they never compare roles themselves. Per-document
permissions are out of scope for milestone 1, so `Can` takes a resource
argument it mostly ignores today. Keep the argument.

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
