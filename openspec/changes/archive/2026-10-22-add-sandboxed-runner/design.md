# Design

Epic: `.beans/chatcms-dupp-sandboxed-build-runner.md`

## What was checked before anything was decided

The means of isolation has to work in two places: on a deployed machine, and
inside the Nix build sandbox, because that is where `nix flake check` runs the
tests. A sandbox that cannot be exercised by the gate is a sandbox nobody
knows works.

Bubblewrap was checked in both. It runs on the host, it blocks the network,
and it runs nested inside the Nix build sandbox. That last one is the reason
the tests in this change run the real sandbox rather than a stub of it.

## Why bubblewrap rather than the alternatives

| Means | Why not |
| ----- | ------- |
| `nix build` per build | Every build becomes a derivation, so the site content has to go into the store, builds cannot be cancelled, and the deployment needs a writable Nix daemon. The cost lands on the wrong side. |
| Namespaces by hand | This is bubblewrap, written again, with the mistakes still ahead of us. |
| Containers | Another daemon to run, to secure and to keep alive, for one short-lived process. |
| Firecracker | The right answer when builds run untrusted code for strangers. Too much machinery for milestone 1, which is why the interface hides the choice. |

Memory is limited by `prlimit` from util-linux rather than by bubblewrap,
which does not do limits. Both are added to the NixOS module's path.

It limits the data segment rather than the address space, and that was
measured rather than assumed. Hugo is a Go program, and a Go program reserves
far more address space than it uses: an address-space limit of one gigabyte
is not enough for Hugo to print its own version, which makes any useful limit
indistinguishable from a broken one. A data limit of 512 MB runs Hugo
comfortably and a limit of 256 MB stops a program allocating four gigabytes,
which is the shape we want.

The kernel enforces that limit against the build, not against the server, so
the only evidence on this side is what the build said on its way out. That is
why the memory limit is reported by reading the log for an allocation failure
and why the spec says a build that says nothing recognisable is reported as an
ordinary failure rather than guessed at. A runner on a machine of its own can
ask the kernel directly, which is one of the things replacing this buys.

## The shape of the interface

A task is inputs, a command, limits and wanted outputs. A result is an exit
code, a log, the collected outputs and what happened to the limits. Nothing in
either names a process, a mount, a path on this machine or bubblewrap.

The test that holds this honest is not a comment: a second implementation that
runs tasks without isolating anything is written in the test package and run
against the same table of cases as the real one. Both have to agree about
everything except isolation. If the interface ever grows a bulge shaped like
bubblewrap, that test stops compiling.

## Two decisions worth stating

**A runner that cannot isolate refuses to run.** The alternative is a runner
that notices the sandbox is missing and runs the build anyway with a warning
in a log nobody reads. That is the worst available outcome: no protection, and
no sign that there is none. The check is done once, when the runner is made,
so refusing is immediate rather than per build.

**Outputs are named, not swept.** A build writes caches, lock files, module
downloads and whatever else it feels like. Collecting the working directory
would carry all of it across the boundary we just built. Naming `public` says
what we came for. A named output that is missing is an error, because a
publish of nothing is worse than a publish that failed.

## What this change does not do

It does not run builds for anything. Nothing calls the runner from an HTTP
path yet: workflows is the epic that gives builds a reason to happen and a
queue to happen in. This change proves a build can be run safely and recorded,
which is what workflows needs underneath it.

One build at a time per project is the queue's business. The job queue already
refuses a second queued job for a subject, and a build is a job, so the
serialising is there rather than here.
