## Why

Building a customer's site means running their build, and a build is code
somebody else wrote. It must not be able to read the other customers' sites,
reach the network, fill the disk, or run forever. Epic `chatcms-dupp`.

It is the precondition for workflows, which is the precondition for
publishing, which is acceptance test 7.

## What Changes

- **Builds run inside a sandbox.** A build sees the files it was given and
  the tools it was named, and nothing else: no other project, no state
  directory, no deploy key, no home directory.
- **A build cannot reach the network** unless it is asked for, which nothing
  in milestone 1 asks for.
- **A build cannot run forever or take the machine with it.** It has a time
  limit and a memory limit, and exceeding either ends it with a log rather
  than a hung server.
- **What a build produced is collected**, by naming the paths wanted, so that
  nothing else it wrote leaks out of the sandbox.
- **Every build is recorded**: what it was for, whether it worked, how long it
  took, and its log.
- **A site is built with the pinned Hugo**, through the runner like anything
  else, so the build a customer gets is the build the gate tested.
- **The interface speaks in inputs, a command and outputs**, with no mention
  of processes or paths on this machine, so that Firecracker micro virtual
  machines can replace the implementation.

## Capabilities

### New Capabilities

- `build/isolation`: what a build can and cannot reach, and what happens when
  it exceeds a limit.
- `build/record`: what is remembered about every build.

### Modified Capabilities

None.

## Impact

- **New code**: `internal/runner`, `internal/sitebuild`.
- **Schema**: one migration adding the builds table.
- **Dependencies added**: `bubblewrap` and `util-linux` at run time, both
  already in nixpkgs and both added to the NixOS module's path.
- **Seams**: this change builds the second of the four seams in
  `docs/ARCHITECTURE.md`, and is the one whose whole purpose is to be
  replaced.
- **Gate**: the tests run the real sandbox, which works nested inside the Nix
  build sandbox; this was checked before the design was settled.

## Assumptions

- **Isolation is required, not best effort.** A runner that cannot sandbox
  refuses to run rather than running without one. A build that quietly ran
  unsandboxed would be the worst of both: no protection and no warning.
- **Memory is limited by the operating system, not measured by us.** The
  limit is applied to the build process; what the log then says depends on
  what the build does when an allocation fails. The guarantee is that it ends
  and that the log is kept, not that the message is always the same.
- **Nothing asks for the network yet.** The flag exists because a build step
  that fetches dependencies is a real need and refusing to model it now would
  mean changing the interface later. No milestone 1 path sets it.
- **A machine that cannot isolate runs no builds, and says so at start-up.**
  The server still starts, because milestone 1 has no path that reaches a
  build yet and refusing to start would block work on machines without
  bubblewrap. Nothing runs unsandboxed either way; builds are simply
  unavailable, and the start-up log says it in those words.

- **One build at a time per project.** Two builds of one site would race over
  the same inputs. The queue already serialises work per subject, which is
  where this belongs rather than in the runner.
