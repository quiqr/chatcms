## Why

Everything built so far works on a site that is already here. Nothing puts one
here. This change imports a real customer's Hugo site, gives the project a key
of its own to reach the remote with, and records how it should be published.
Epic `chatcms-zytv`.

It is acceptance test 2, "import the customer's Hugo repo with its Quiqr model
using a deploy key", and it is the precondition for publishing.

## What Changes

- **A project gets a key of its own.** An SSH key pair is generated per
  project. The public half is shown so an implementer can give it to the git
  host; the private half is encrypted before it is stored and is never shown,
  returned or logged.
- **Secrets are encrypted at rest.** A master key, read from a file named by
  configuration, encrypts every stored secret. Nothing readable is written to
  the database, and the master key is never in the Nix store.
- **A site is imported from one of three places.** A git repository, a site
  exported from the Quiqr desktop app, or a Quiqr template.
- **What arrived is inspected and reported plainly.** Whether it is a Hugo
  site, and whether it carries a Quiqr model. A repository with no model is
  not a broken project: it is a project that needs one, and says so.
- **The connection can be tested, before and after.** An implementer can ask
  whether the key reaches the remote, and get an answer that says what to do
  rather than what failed.
- **How a site is published is recorded.** To the main branch or to one of its
  own, and either the source as it is or the built result.
- **The onboarding screen exists**, with its step list and its three sections,
  as the designs draw it.

## Capabilities

### New Capabilities

- `tenancy/secrets`: how a secret is stored so that a copy of the database is
  worth nothing.
- `onboarding/import`: how a site arrives, what is checked about it, and what
  an implementer is told.
- `onboarding/publishing-target`: where a site is published and in what form.

### Modified Capabilities

- `tenancy/organisations`: a project now carries where it came from and how it
  is published.
- `sync/repository`: a working copy now reaches its remote with the project's
  own key rather than one shared by the server.

## Impact

- **New code**: `internal/secrets`, `internal/onboarding`, their handlers, and
  the onboarding screen.
- **Changed code**: `internal/gitsync` takes a key per site rather than one
  for the server; `internal/tenancy` carries the new project fields.
- **Schema**: one migration adding the key, the source and the publishing
  settings to a project.
- **Dependencies added**: none beyond `golang.org/x/crypto`, which is already
  here for argon2id and carries the SSH key marshalling.
- **Seams**: `internal/secrets` is new and is the thing a hardware key store
  would replace. `gitsync.Repo` is unchanged in shape; only what it is handed
  changes.

## Assumptions

- **A desktop-app export arrives as a directory on the server.** The desktop
  app keeps a site as a folder. Uploading one through the browser needs the
  file handling the media epic builds, so for now an implementer places the
  folder where the server can read it and names the path. The import then
  makes it a repository and pushes it, which is the part that matters.
- **Templates are repositories.** Starting from a template is importing a git
  repository whose address ChatCMS knows, so there is one import path rather
  than two.
- **The private key is written to disk while the server runs.** Git needs a
  file to read a key from. It is written with owner-only permission into a
  directory made for the purpose when the server starts, and the directory is
  removed when it stops. The design records why this is the least bad option.
- **Nothing rotates a key yet.** A project's key can be generated and
  replaced; scheduled rotation, and the host-side bookkeeping it needs, is not
  milestone 1.
- **The connection test cannot distinguish every failure.** A host that
  refuses a key and a host that is down look different; a host that is
  reachable but has not been given the key yet is the case that matters, and
  that one is named exactly.
