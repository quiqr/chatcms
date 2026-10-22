## Context

`internal/gitsync` can clone, save and exchange; `internal/tenancy` has
projects; `internal/model` can read a model. What is missing is the step that
turns a customer's existing site into one of those projects, and the key it
uses to reach the remote. See `proposal.md`.

## Goals / Non-Goals

**Goals:**

- A copy of the database is worth nothing to whoever takes it.
- An implementer who has done this once can do it again in minutes, and knows
  what to do when it does not work.
- Importing twice from the same site, once as a repository and once as an
  export, gives the same project.

**Non-Goals:**

- Uploading an export through the browser. See the assumption in
  `proposal.md`.
- Key rotation on a schedule, and the host-side bookkeeping it needs.
- A GitHub App, which milestone 2 holds and which would replace keys entirely.

## Decisions

### Ed25519, generated here, never shelled out for

Keys are generated with Go's own `crypto/ed25519` and marshalled into the
OpenSSH formats with `golang.org/x/crypto/ssh`. No `ssh-keygen` subprocess,
which means no temporary file holding a private key during generation and one
less binary in the deployment.

Ed25519 rather than RSA because every git host worth using has accepted it for
years, the keys are short enough to paste, and there is no key size to choose
wrongly.

### Secrets are sealed with AES-256-GCM under one master key

One master key, read from the file `CHATCMS_SECRET_KEY_FILE` names, encrypts
every secret with AES-256-GCM and a fresh nonce each time. The nonce is stored
beside the ciphertext. Authentication is what makes a tampered row fail to
decrypt rather than decrypt into something.

The key is a file rather than an environment variable because the NixOS module
can point at a path outside the store, and because an environment variable
ends up in a process listing and in crash reports.

Alternative considered: a key per project, wrapped by the master key. More
moving parts for the same property while there is one master key to lose.

### The private key is written to disk while the server runs, and nowhere else

Git reaches a remote by running `ssh -i <file>`. There is no way to hand it a
key in memory without writing an agent, and an agent is a long-lived process
holding every key for every project.

So: at start-up the server makes a directory with owner-only permission inside
its state directory, and removes it on shutdown. A key is decrypted into a
file there the first time a project needs it, with owner-only permission, and
stays for the life of the process.

This is the least bad option and it is worth naming what it costs: anyone who
can read that directory can read every project's key. The protections are the
file mode, the directory mode, and the fact that the process runs as its own
user, which the deployment module already arranges.

### Importing is one path with three doors

A repository is cloned. An export is copied into a new working copy, committed
and pushed. A template is a repository whose address ChatCMS knows, so it is
the first case.

After that, every source is the same thing: a working copy with a remote. The
inspection, the model check and everything downstream do not know which door
was used, which is why an export and its repository give the same project.

### What was imported is inspected, not guessed

Hugo is detected by its configuration file, under any of the names and places
Hugo accepts. A model is detected by `internal/model`, which already knows
where to look and already reports its own errors well.

A site with no model is not an error. It is a site that needs one, and saying
that is more useful than refusing the import, because an implementer often
imports first and writes the model second.

### Testing the connection reports the case that matters

Three outcomes: the remote took the key, the remote is there and refused the
key, or the remote did not answer. The middle one is the case that happens on
every first setup, and it is the one whose message says exactly what to do:
the key is shown again, with where to paste it.

Distinguishing refused from unreachable is done on the error git returns,
which is a heuristic. When the heuristic cannot tell, it says the remote could
not be reached, which is the safer thing to be wrong about: it sends the
implementer to check the address rather than to re-paste a key that was fine.

## Risks / Trade-offs

- **One master key.** Losing it loses every deploy key, and every project then
  needs a new one on the git host. Mitigated by the keys being replaceable and
  by the deployment epic documenting the key as the thing to back up.
- **Decrypted keys on disk for the life of the process.** Named above. The
  alternative is an agent, which is a bigger surface for the same risk.
- **Detecting Hugo by its configuration file** is a guess that a site with an
  unusual layout can fail. Mitigated by saying what was looked for, so an
  implementer can see why rather than wondering.
- **An export is read from a path on the server**, which means an implementer
  needs access to the machine. That is true of this milestone's deployment
  anyway, and the media epic removes it.

## Migration Plan

One migration adds the key, the source and the publishing settings to a
project. Existing projects, of which there are none deployed, would get a key
the first time one is asked for.

## Open Questions

- Whether a project should be able to publish to more than one place. The
  designs show one. Adding a second is a table rather than a column, and
  nothing in milestone 1 needs it.
