## Why

The briefing's section 6 lists nine acceptance tests and calls them the
definition of done. Until they run by themselves, "done" is a reading of the
code rather than a fact about the product. Epic `chatcms-va5c`.

Bean: `.beans/chatcms-va5c--end-to-end-acceptance-suite.md`.

## What Changes

- **Nine specs, one per acceptance test**, in `test/e2e/`, driving the real
  browser against the built binary.
- **Every spec runs twice**, at 1440 px and at 390 px, and the phone pass adds
  the two phone rules: no sideways scrolling and nothing under 44 px.
- **A harness that owns everything it needs**: its own PostgreSQL, its own
  bare remote, its own seeded site, its own server, torn down afterwards.
- **The `e2e` check in `flake.nix`**, so the suite runs in the gate with the
  browser, Hugo, git and PostgreSQL all pinned.
- **A capture of every screen at both widths**, next to the drawings they came
  from, with the differences written down in `docs/design/COMPARISON.md`.
- **A scripted run against a real site**, `scripts/acceptance-against.sh`, so
  the suite can be pointed at the customer's own repository.
- **The phone editor gets a preview**, because acceptance test 5 is "works on
  a phone" too and the phone layout had no way to show one.

## Capabilities

### New Capabilities

- `acceptance/coverage`: which tests exist, what each proves, and the two
  rules every phone pass adds.
- `acceptance/harness`: what a run owns, what it borrows, and what it is never
  allowed to do.

### Modified Capabilities

None. The suite asserts the behaviour other changes specified; it does not
change what that behaviour is.

## Impact

- **New**: `test/e2e/` (specs, harness, fixtures, screen capture),
  `nix/e2e.nix`, `scripts/acceptance-against.sh`,
  `docs/design/COMPARISON.md`.
- **Changed**: `flake.nix` gains the `e2e` check; `cmd/seed` grows a site that
  is worth testing against; the phone editor gains an Edit and Preview switch;
  a handful of controls grow to 44 px on small screens. The integration
  harness now stops its PostgreSQL when the test binary ends: writing down
  that a run leaves nothing behind made it plain that the older harness left
  one cluster per package running until the machine was restarted.
- **Bugs this found**, each fixed in the code rather than worked around in the
  test: the editor route did not match a page whose key holds a slash; the
  sync screen threw the answer away after bringing changes in; preview links
  and live reload pointed at a port nobody outside the machine could reach; a
  master key of raw random bytes was refused whenever its last byte happened
  to be a newline, which is one start in a hundred.

## Assumptions

- **A test owns its world.** Every spec gets its own database, remote and
  server, so the order they run in cannot matter and a failure is never
  somebody else's leftovers. It costs a few seconds per spec and buys a suite
  that says the same thing on the tenth run as on the first.

- **No sleeps.** Waiting for a wall clock is waiting for the wrong thing: it
  is slow when the machine is fast and wrong when it is slow. Every wait here
  is for the thing itself, a response, an element, a state, with a deadline.

- **The drawings are compared by reading, not by pixels.** A design is a
  drawing in a design tool; the product is a browser rendering a different
  font engine. A pixel comparison would report a thousand differences that
  mean nothing and miss the one that matters. So the screens are captured and
  the differences are written down, which is what the bean asks for.

- **The suite runs against a fixture site in the gate, and can be pointed at a
  real one.** The customer's repository is not something a build sandbox may
  reach, and a gate that needs the network is a gate that goes red when
  somebody else's server has a bad morning. `scripts/acceptance-against.sh`
  runs the same specs against any repository given to it, which is how the
  run against the customer's site is done, by hand, before a release.

- **The phone editor needed a preview before test 5 could pass on a phone.**
  Fixing the product rather than lowering the test is the point of having the
  test.
