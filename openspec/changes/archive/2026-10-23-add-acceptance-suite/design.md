# Design

## What a run owns

A spec gets a `Harness`: a PostgreSQL in a temporary directory, a bare remote
on disk, a seeded database, and the server built by this flake. All of it is
removed afterwards, including after a failure, because the teardown is in the
fixture rather than at the end of the test body.

That is deliberately wasteful. Starting PostgreSQL for every spec costs a few
seconds each; sharing one would save a minute across the suite and buy back a
whole class of failures where a spec passes alone and fails in company. The
suite is for telling the truth about the product, not for being quick.

The remote is a bare repository on disk rather than a git host. It behaves
like a remote in every way the product cares about: it is cloned from, pushed
to, and read back by a test that wants to know what really arrived. Nothing
in the product knows the difference, which is the point.

## Waiting

Nothing in the suite sleeps. Waiting for a wall clock is waiting for the
wrong thing: too slow when the machine is fast, too short when it is loaded,
and silent about which it was. Every wait is for the thing itself:

- the server, by asking `/api/health` until it answers, with a deadline;
- a save, by waiting for the response that writes it;
- a preview, by waiting for the words to appear in the frame;
- work arriving from elsewhere, by waiting for the screen to say so.

`saves(page, change)` is the one worth naming. It waits for the write to come
back before it looks at the saved indicator, because the indicator is the
thing under test and reading it too early is how a suite learns to pass by
accident.

## Both widths in one suite

Rather than a phone tour of its own, every spec runs twice, and the phone pass
ends with `worksOnAPhone`. That is what keeps acceptance test 9 honest: it is
not a list of screens somebody remembered to add, it is every screen every
other test already walks. The handful of screens no other spec lands on get a
spec of their own that does nothing but visit them and apply the same two
rules.

The two rules are the ones from the design: the page may not scroll sideways,
and nothing that can be pressed may be under 44 px. Visually hidden inputs are
exempt, because a 1 px input behind a styled label is how a file picker is
made and the label is what the thumb lands on.

## The drawings

`screens.shots.mjs` captures each screen at both widths. It is not run in the
gate and it asserts nothing about pixels. A drawing made in a design tool and
a browser rendering a different font engine will differ in a thousand places
that mean nothing, and a comparison that cries wolf is a comparison nobody
reads. So the screens are captured, a person reads them beside the drawing,
and `docs/design/COMPARISON.md` says what differs and what was decided about
it. That is what "listed rather than silently accepted" has to mean here.

## A real site

The gate runs against the fixture. A sandbox may not reach the network, and a
check that depends on a git host is a check that goes red when somebody else
has a bad morning.

`scripts/acceptance-against.sh` runs the same specs against any repository,
by giving the seed `CHATCMS_SEED_REMOTE` so it clones that site instead of
making the fixture one. The specs do not change. A spec that names a page the
real site has not got fails on the page, which reads as "this site has no
such page" rather than as a broken product, and that is the right answer for
a run whose purpose is to find out what a real site does.

## Fonts in the build sandbox

The suite ran green on a laptop and failed every sign-in in the Nix sandbox.
The server was answering, the page was rendering, the click was landing, and
nothing was posted. What was happening: Chromium had no fonts and no
fontconfig, so the first time it had to pick one it hit

    FATAL: SkFontMgr_FontConfigInterface.cpp: Not implemented.

The renderer dies, but the page is already drawn, so what a test sees is a
screen that looks right and ignores every keystroke. The form then refuses to
submit because its required fields are empty, which is why nothing reached
the server and why no error appeared anywhere.

The fix is one line: `FONTCONFIG_FILE` pointing at a generated config with
DejaVu and Liberation in it. It is written down here because the symptom
looks nothing like the cause.

## What the suite found

Four product bugs, each fixed in the product:

- The editor route did not match a page whose key holds a slash, so every
  page below the top level was a not-found.
- The sync screen threw away the answer from bringing changes in, so the
  screen still said there was work to fetch after it had been fetched.
- Preview links and live reload pointed at the port Hugo listens on, which
  nothing outside the machine can reach. Hugo is now told the address the app
  serves it under.
- Two controls were under 44 px on a phone, and the phone editor had no way
  to see a preview at all, which is half of acceptance test 5 on a phone.

Finding these is the argument for the suite. None of them would have shown up
in a unit test, and all four are the kind of thing a customer meets in the
first ten minutes.
