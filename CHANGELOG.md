# Changelog

All notable changes to Quiqr ChatCMS are documented here.
The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/)
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Editing. A page opens as a form built from the site's own Quiqr model, every
  field type is drawn, and what is typed is saved by itself with no Save
  button anywhere. Saving writes the file back byte for byte apart from what
  changed.
- A content browser: the site's pages and collections, with a mark on anything
  waiting to be published, search, and making, renaming, copying and removing
  pages.
- A picture library: add, search, describe for screen readers, see which pages
  use a picture, and pick one from a form field. What a picture is comes from
  reading it, never from what the file was called.
- Publishing. A review of what changed, said as pages and fields with the
  words that differ marked, an optional note, and then the publish itself, to
  the main line or a line of its own, as source files or as built pages.
- History, and going back. The timeline reads in plain language, and going
  back to an earlier version is itself something that can be undone, because
  it makes a version rather than removing any.
- Live preview of a site while editing it, reached through the app so that
  unpublished work is never readable by anybody who may not see the project.
- Declarative workflows in `quiqr/workflows/`, so a site can say what happens
  when it is saved, previewed or published.
- Site builds run isolated, with a time and memory limit, and no network.
- Sync status in the five states the designs ask for, and a question per field
  when two people changed the same thing.
- A NixOS module and a virtual machine test that boots it, runs the migrations
  from an empty database, and asserts the master key never reaches the Nix
  store. See `docs/DEPLOYING.md`.
- The nine acceptance tests run by themselves, in a real browser against the
  built binary, at a desk width and at a phone width, as part of
  `nix flake check`. `scripts/acceptance-against.sh` runs the same tests
  against a real site.

- Project scaffold: OpenSpec, beans, Nix flake, ship gate and the milestone 01 backlog.
- The server runs. `chatcmsd` listens, serves a JSON API, reports its health at
  `/api/health`, and stops gracefully without dropping a request in flight.
- The database schema is created and upgraded by the server itself on start-up,
  safely even when two instances start at the same time. A server older than the
  database it is pointed at refuses to start rather than serving against a schema
  it does not understand.
- A deployment is one binary. The browser application ships inside `chatcmsd`, so
  there is no separate directory of assets to copy or keep in step.
- Configuration values can be supplied from files rather than environment
  variables, so passwords and keys stay out of the Nix store.
- Quiqr models are read the way the desktop app reads them: a base file in
  YAML, TOML or JSON, plus includes, per-item include folders and partials,
  assembled in the same order and with the same precedence.
- All twenty-seven field types of the desktop app are understood, and a field
  type this version does not know is kept as a placeholder rather than dropped,
  so a model written for a newer Quiqr still opens.
- An unusable model says which file and which part of it is wrong.
- The pages and collections a content creator browses, with item counts and
  the model's own labels, are derived from the model itself.
- Editing a site makes a version of it, attributed to the person who made the
  edit, so history reads as a list of who changed what.
- Changes made to a site outside ChatCMS arrive on their own, without anyone
  asking for them, and are woven in underneath the work in progress.
- When two people change the same page, the site says so and offers both
  versions to choose between. Neither version is lost, and nobody ever opens a
  page full of machine-written markers.
- Publishing, syncing and anything else that has to talk to the outside world
  happens in the background, so a slow or unreachable host never stops someone
  editing. Work that fails is tried again, and work that never finished is
  picked up after a restart rather than lost.
- Nothing the system records about a failure carries a password or a key.
- A page saved through ChatCMS comes out of the other end spelled exactly as
  the Quiqr desktop app would have spelled it, down to which strings are
  quoted and where a long line wraps. Adopting ChatCMS does not rewrite a
  repository.
- Editing one field changes one line. Values the model does not mention are
  left exactly where they were, and a field the model declares but the file
  lacks appears in the place the model implies.
- A page whose front matter is TOML or JSON can be read but not saved, and
  says so. The desktop app writes TOML in a layout it cannot read back, so
  matching it would produce files nothing can open.
- The interface has its own look: the Quiqr colours and typefaces, the buttons,
  badges and menus the designs call for, and the frames they sit in. Phones get
  their own frame with a bottom tab bar rather than a squeezed desktop.
- The typefaces are served by ChatCMS itself, so no reader's browser is asked
  to contact anyone else.
- What an editor can read is checked by the build: a version control word in a
  screen a content creator sees fails it, as does a colour written outside the
  palette or a control too small to tap.
- People can sign in, stay signed in on a device if they ask to, and sign out.
  Signing out ends the session on the server, not just in the browser.
- An invitation link sets a first password. It works once and then stops
  working, as does a password reset link, and resetting a password ends every
  session that person had.
- Nothing ChatCMS says reveals whether an email address belongs to an account,
  whether someone is signing in, asking for a reset, or using a link that has
  already been used.
- Repeated guessing is slowed down, counted both against the address being
  tried and against whoever is trying.
- Until there is a way to send email, an invitation or reset link is written to
  the server log for an implementer to pass on by hand, and is never returned
  to the browser.
- There are organisations with projects inside them, and four roles: owner,
  implementer, editor and viewer, with exactly the powers the design gives
  each.
- An owner invites someone by email address and role. Someone new gets an
  account and a link that sets their first password; someone already known
  simply gains the role.
- A viewer can read and cannot change anything, and this holds whether the
  button is on the screen or not.
- What someone may not do is left out of the interface rather than shown greyed
  out, because the server tells each person what they may do.
- A project belonging to an organisation you are not in looks exactly like a
  project that does not exist, so the addresses of other customers' sites
  cannot be discovered by guessing.
- The project list and the people and roles screen are built and read real
  data.
- A customer's existing site can be brought in: from where it already lives in
  git, from a folder exported by the Quiqr desktop app, or from a template.
- Each project gets a key of its own for reaching where its site lives. The
  half you paste into your git host is shown with what to do with it; the other
  half is encrypted before it is stored and is never shown, returned or logged.
- After bringing a site in, ChatCMS says what it found: whether it is a Hugo
  site and whether it has a Quiqr model. A site without a model is not broken,
  it just needs one, and the message says where to put it.
- The connection can be tested at any time, and the answer says what to do
  next rather than what went wrong.
- Where and in what form a site is published is recorded per project: to its
  main line or to one of its own, and either the site's files for your own
  pipeline to build or the result of ChatCMS building it.
