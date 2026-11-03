# Tasks

## 1. What is pinned

- [x] 1.1 `scripts/reference-pins.env` holds the Quiqr commit, the kitchen
      sink commit and the two ports, each with a line saying what moving it
      means. Verified by `bash -n` on the script that sources it and by every
      later task reading its values rather than repeating them.
- [x] 1.2 The script prints those pins plus the node, Hugo and embgit the old
      app's dev shell gave it, every start, before anything else. Verified by
      running `up` twice and seeing the same block both times.

## 2. The application

- [x] 2.1 `scripts/reference.sh up` clones `quiqr/quiqr-desktop` at the pinned
      commit into `.reference-state/app` and does every build and run step
      inside that checkout's `nix develop`, with `LD_LIBRARY_PATH` carrying
      libstdc++ so the prebuilt native modules load. Verified by running it on
      a machine with no checkout and seeing the build finish.
- [x] 2.2 It builds and starts the standalone edition on the pinned port, with
      `QUIQR_CONF_DIR` under `.reference-state/conf`. Verified by the address
      it prints answering with the old app's own interface.
- [x] 2.3 It makes a user whose address and password are printed, and clears
      the forced password change in `users.json`. Verified by signing in with
      the printed credentials and reaching the site list without being asked
      to change anything.
- [x] 2.4 A second `up` with the state already good starts the server without
      fetching or rebuilding. Verified by timing it: no npm output, no clone.

## 3. The site

- [x] 3.1 The kitchen sink is cloned at its pinned commit into
      `.reference-state/site` and mounted by writing
      `conf/sites/<key>/config.json`. Verified by the site appearing in the
      old app with its pages and collections.
- [x] 3.2 The script reports what the mounted model declares: every field type
      in it, counted. Verified against the 26 types the template is known to
      declare, and naming `easymde` as the drawable type it omits.
- [x] 3.3 An edit made in the old app shows up as a diff in
      `.reference-state/site`. Verified by changing one value and reading
      `git -C .reference-state/site diff`.

## 4. Giving the machine back

- [x] 4.1 `stop` ends every process the script started and frees both ports.
      Verified by `stop` followed by both ports being free.
- [x] 4.2 `fresh` throws the checkout, the installed packages, the
      configuration and the site away, and the next `up` rebuilds from
      nothing. Verified by `fresh && up` on a machine that had a working
      instance.
- [x] 4.3 `.reference-state/` is ignored, so a run leaves the repository
      clean. Verified by `jj st` being empty after `up`.

## 5. One field, all the way round

- [x] 5.0 A recorded case under `test/fixtures/reference/string-in-front-matter/`:
      the model, `dataFields.md` before, `dataFields.md` after, and the edit
      (`simpleDataFields`, field `string`, set to a known value), made by
      editing it in the old application and committing what it wrote.
      Verified by the recorded "after" differing from "before" in exactly the
      lines the edit touches.
- [x] 5.0.1 The recording carries the versions that produced it, as
      `test/fixtures/content/golden/index.json` does. Verified by reading it.
- [x] 5.0.2 A Go test loads the recording, performs the same edit through
      `internal/forms`, and asserts byte equality with the recorded "after",
      final newline included. Verified by `go test ./internal/forms/`, which
      needs no network and nothing running.
- [x] 5.0.3 The same test asserts every value the edit did not name is
      unchanged, so a case cannot pass while a neighbour was rewritten.
      Verified by making it fail on purpose: change a neighbour in the
      recorded "after" and watch the test complain about that value by name.
- [x] 5.0.4 If ChatCMS and the recording disagree, the disagreement is
      written up in the change's summary and a decision taken. The test is
      not loosened to make it pass.

## 6. Proving the harness

- [x] 6.1 A smoke run from nothing: clone, build, start, mount, sign in, edit,
      see the diff, record the case, stop, fresh. Done by hand and written up
      in the change's summary, because it needs the network and cannot be a
      gate check.
- [x] 6.2 The script fails loudly, with a sentence naming what it expected,
      when a pinned commit cannot be fetched, a port is taken, or `users.json`
      is not the shape it knows. Verified by forcing each of the three.
- [x] 6.3 No check is added to `flake.nix` for the script, deliberately: it
      reaches the network, so it is a generator of evidence rather than a
      check, like `scripts/generate-goldens.sh`. The test it produces is an
      ordinary `go test` and is therefore already in the gate. Verified by
      `nix flake check` running the new test and not the script, and by the
      reason being written down in task 7.1 rather than left to be
      rediscovered.

## 7. Writing it down

- [x] 7.1 `docs/TESTING.md` gains a section beside the goldens one: what the
      reference instance is, that it is an oracle for what a field does rather
      than for how bytes are written, why making a recording is not in the
      gate while checking one is, what `up`, `stop` and `fresh` do, and how to
      add the next field type. Verified by a reader who has never run it being
      able to do both.
- [x] 7.2 `docs/TESTING.md` says plainly which field types are covered by a
      recording: one. A green gate must not be read as more than it is.
- [x] 7.3 `CHANGELOG.md` under `[Unreleased]`, in the section for tooling.
- [x] 7.4 The bean `chatcms-37w8` is closed with a summary that records what
      the first run actually showed about the kitchen sink in the current old
      Quiqr, including anything the template does that the application no
      longer does.

## Found on the way

Two things the reference said on its first recording, which is what it is for:

- The old Quiqr writes a hidden field's default into the file on any save.
  Named as a divergence in `case.json` with its reason; ChatCMS does not put
  a value in a repository that nobody typed.
- A document whose front matter ends with a blank line does not survive a
  no-op round trip in ChatCMS: the blank line goes and two newlines arrive at
  the end of the body. The old Quiqr does the same two things, so the
  recorded case passes and a customer's repository gets the same bytes either
  way. `content/serialisation` nonetheless claims a no-op round trip changes
  nothing, which is not true for a real file from a real template. Recorded
  as `chatcms-8d0w` rather than fixed here, because fixing the writer is a
  behaviour change and this change is a harness.
