# Changelog

All notable changes to Quiqr ChatCMS are documented here.
The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/)
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

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
