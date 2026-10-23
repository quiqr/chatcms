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
| `internal/publish`     | What changed as pages and fields, the publish itself, and its record          |
| `internal/content`     | What a site holds, one page as a form, and writing an edit back               |
| `internal/media`       | The pictures a site carries, what uses them, and small copies for a grid      |
| `internal/sitesync`    | How a site stands, and what two people changed, as pages and fields           |
| `internal/history`     | The timeline in plain language, what a version holds, and going back          |
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

**`internal/content`** turns a page into the controls an interface draws, and
writes an edit back through `internal/forms`, which is what keeps the file
byte for byte as it was apart from what changed.

The browser is given controls with values and a path for each, rather than the
model. Working a path out means knowing that a grouping field may or may not
group its data, which the binder already knows; sending the model would mean
writing those rules twice, in two languages, and having them drift.

A kind of field this version cannot draw is shown as plain text, saying so,
rather than left out. Leaving it out would make a page look emptier than it is
and lose what was in it on the first save.

A page's own text is editable only where the model declares a field for it,
which is how the reference implementation works. Where it does not, the text
is shown and an attempt to write it is refused with a reason rather than
quietly dropped.

It also answers what a site holds. A page's file is named after its title the
way the old Quiqr names it, which is why renaming moves the file and rewrites
the title together: to a content creator those are one act. Whether an item is
waiting to be published comes from the same question the publish review asks,
so the two cannot disagree.

**`internal/media`** keeps the pictures. What a picture is comes from reading
its bytes, never from its name: a name arrives from outside, and a misnamed
file would be published as a broken picture on a customer's site.

What uses a picture is worked out by reading the content each time rather than
from an index, because an index would be wrong the first time anybody edited a
page outside ChatCMS, which this product supports on purpose.

A small copy is named after a hash of the bytes it was made from, which makes
the thumbnail directory a cache in the strict sense: a changed picture gets a
new name by itself, and deleting the whole directory costs only the work of
making them again from the site in git.

**`internal/sitesync`** turns what the repository engine reports into
something a content creator can act on. The engine answers in files; this
turns a file into a page and a page into questions about its fields, using the
same model and form loader the editor uses. Nothing about conflicts moved into
the engine and nothing about git moved out of it.

Only a file that holds words gets field questions: a picture parses as a
document whose body is the binary, and offering that as two answers would be a
screen of mojibake. A field nobody answers about keeps what was written here,
because doing nothing should never throw away the work of the person doing it.

Technical detail is left out of the answer for somebody who may not see it,
not merely unrendered: a screen that hides what it was sent is one view-source
away from being wrong about who may see what.

**`internal/history`** lists what happened to a site and puts it back.

Going back makes a new version holding an earlier version's content. It
removes nothing, so the history keeps everything and going back is itself
reversible: the version somebody moved away from is still in the timeline.
Unpublished work is set aside rather than lost, which is accurate rather than
soothing: the version holding it stays in the history and going back to that
version brings it out again.

A version's title comes from the note somebody wrote when publishing, or its
own message, or a summary of what changed. The third exists because some
messages are written by machinery, and those are both meaningless to a
content creator and full of words the copy rules forbid.

Showing what one version changed is the same question as showing what is
waiting to be published, asked of a different pair of trees, so it reuses
`internal/publish` through an adapter rather than being written twice.

**`internal/publish`** turns what differs from the published site into pages
and fields, runs the publish workflow, and records what happened.

It asks the repository three things the interface learned for it: what
differs, what the live site holds for a file, and to publish to a named line
either as source or as built pages. All four methods speak in sites and
lines.

Publishing built pages never touches the working copy. The tree is built with
git's own plumbing in a temporary index, so somebody can keep typing while
their site goes out. That is what makes "you can keep editing" true rather
than hopeful.

A failure carries two things, separate fields the whole way from the workflow
outcome to the screen: a sentence for whoever pressed the button, with no git
word in it, and the raw output. `policy.Can` with `see-technical-detail`
decides who gets the second, asked in the service, which empties the field
before it leaves.

**`internal/workflows`** reads what a site says happens on save, preview and
publish, from YAML under `quiqr/workflows/`, and runs it:

```yaml
on: publish
steps:
  - name: Checking pages and links
    action: check
    run: ./scripts/check-links.sh
  - action: build
  - action: publish
```

A key the format does not define is a mistake, not something to ignore, and
every error names the file and the step. A trigger the site says nothing
about gets a default worked out from the project's publishing settings, so an
imported site publishes correctly before anybody writes a workflow.

Which steps are contained, and which are not:

| Step | Runs code the site wrote | Sandboxed |
| --- | --- | --- |
| `check` | Yes, the command in `run` | Yes |
| `build` | Yes, the site's templates and content | Yes |
| `publish` | No | No; it needs the network and the deploy key |

Saying "every step runs in the sandbox" would be a comfortable untruth. The
sandbox bounds what the site's own code can do. Publishing is ChatCMS acting
on the project's settings, with no site input to contain.

A check gets a fixed, pinned toolbox, named by a setting and supplied by Nix.
A site cannot add to it. Letting a workflow name its own dependencies means
fetching software from the internet in the middle of a publish, which is the
shareable-actions idea the briefing puts out of scope.

A failure carries two things, kept apart in the type so the second is hard to
leak: a sentence for the person who pressed the button, with no git word in
it, and the raw output on the step. This package never decides who may see
the second one; `policy.Can` with `see-technical-detail` does, where it is
rendered.

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
