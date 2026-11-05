# Changelog

All notable changes to Quiqr ChatCMS are documented here.
The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/)
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Signing in as each role in one press during development. Under
  `npm run dev` the sign-in screen offers a button for the platform admin,
  the implementer, the editor and the viewer, and each goes through the
  ordinary password sign-in. The seed now gives the editor and the viewer the
  same passphrase as the other two, so all four can sign in. A production
  build carries none of it, and the frontend check fails when the seeded
  passphrase or addresses reach the bundle.
- A reference instance of the old Quiqr, set up by one script, and one field
  carried all the way round the circle. `scripts/reference.sh` clones the old
  application and the kitchen sink template at pinned commits, builds the
  standalone edition inside that checkout's own dev shell, and mounts the
  template as a site, so a question about what a field does is answered by
  running the application rather than by reading its source. `record` keeps
  what an edit did as files, and a test in the gate replays the same edit and
  asserts byte equality, plus that no value the edit did not name moved. One
  field type is recorded so far: a string, in front matter surrounded by
  twenty values of other kinds.

- A whole site duplicated, moved, downloaded or archived. A copy carries the
  pages, the media and the content model, takes a name of its own and
  publishes nowhere unless that was asked for. Moving a site to another
  organisation takes its people along and needs both organisations. A site
  comes down as a zip of its files, with nothing of how versions are kept in
  it, and an archived site stops counting against the organisation's limit
  until it is brought back.

- Every user account, and what may be done to one. An account is invited,
  active or blocked rather than on or off, and the platform admin can put
  right a name or an address typed wrong, send a password reset, add somebody
  to an organisation, make or unmake a platform admin, block or unblock, and
  delete an account. Deleting takes the account and its memberships; what
  that person wrote stays on the sites they worked on, in their name. The
  list is searched, filtered and paged by the server, a page at a time.

### Changed

- The platform admin has an area rather than a page. A dark shell that says
  where you are, every organisation with who administers it, what it holds
  against its limits and when it was made, and one organisation opened on a
  screen with its details, its limits, its sites, its members and its
  removal. Searching and the filters are answered by the server. Removing an
  organisation now takes its sites and memberships with it, behind a dialog
  that lists what goes and asks for the organisation's name to be typed;
  nothing in git is touched, as before.

### Added

- An organisation has screens of its own. Whoever administers one reaches
  them from the switcher and sees what the organisation holds: its sites as a
  list, who is in it with the sites each person reaches, and how full it is
  against its limits. Inviting somebody and taking them out happen there, and
  taking somebody out of an organisation now takes their roles on its sites
  with them. The People and roles screen on a project keeps the project; the
  organisation half of it moved to the organisation.

### Added

- Limits per organisation. An organisation may be allowed a number of members
  and a number of sites; both are optional, both are set by the platform
  admin, and both are refused before anything runs. The refusal names the
  limit and the numbers and says what can be done about it, and a screen that
  would be refused says so before it is used rather than after. A pending
  invitation holds a seat. Lowering a limit below what is already there
  removes nothing: it stops what is new until usage is under it again.

### Changed

- Three levels of role, from the design hand-off of 2 November. A platform
  admin holds the installation, an organisation admin holds one organisation,
  and a project has implementers, editors and viewers. The project role owner
  is gone: billing, deleting a project and making one belong to whoever
  administers the organisation, and an organisation admin reaches every
  project in theirs as its implementer. A migration moves every project owner
  to implementer and every organisation owner to admin, so nobody loses the
  work they were doing.

### Fixed

- Switching organisation. The switcher in the header listed every
  organisation somebody belongs to and did nothing when one was chosen; the
  only way across was to type the address. Each one is now a way to that
  organisation's projects, and the one being looked at is marked.

### Added

- Dutch, and one place for every string. Every string the interface says now
  comes from `web/src/strings/`: the English is the catalogue and the type,
  and Dutch is a file holding exactly its keys, so a missing string fails the
  build rather than appearing in the wrong language on somebody's screen. A
  check fails when a string is written in a screen, the way the interface
  check already fails on a colour written by hand, and it reads the catalogue
  in both languages so that no version control word can come back in through
  a translation. A language is chosen on the preferences screen beside the
  appearance, is kept with the account, takes effect at once and follows to
  another machine; somebody who has chosen nothing is read to in the language
  their browser asks for. Dates and times follow it. A refusal from the server
  carries a key as well as its English reason, so the screen says it in the
  language whoever is reading chose.

- Acting as somebody else, visibly. Whoever runs the installation can be a
  customer for a while: their screens, and what they may do, so that "I
  cannot see the publish button" can be answered by looking. It announces
  itself the whole time. A band across every screen says who is being acted
  as and offers the way out, the session stays the administrator's and says
  who is really there, and everything saved is attributed to both: the
  site's own history and its git repository read "Pim Snel as Marieke de
  Vries". Beginning and ending are both written down. Somebody who
  administers the installation cannot be acted as, and neither can an
  account that is off.

- Every person on the installation. Whoever runs it now sees who is here:
  each person's name, address, whether their account is on, when they were
  last here, and every organisation and site they can reach, read in one
  query rather than one per person. The list can be searched by name or
  address, each person opens on a screen of their own where their account is
  turned off or on again, and a person can be made from an address and a
  name. Making one sends the same way in as every other invitation and sets
  no password: an administrator never types somebody else's.

- Administering an organisation. Whoever runs the installation opens an
  organisation from the administration screen and sees everything in it at
  once: its name, who is in it with what role, and the sites it holds. The
  name can be changed, people can be put in or taken out, and an organisation
  with no sites left can be removed after a confirmation that says what stays.
  One that still holds sites is not removed; the refusal names them. Nothing in
  git is ever touched by any of this.

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
- A dark appearance, and following whatever your device asks for, which is
  what you get without choosing anything. The choice is kept with the
  account, so it is the same wherever you sign in.
- Every date and time in the interface is now written one way, in one place,
  instead of four screens each deciding for themselves.
- An account has a picture, which is shown wherever that person is. It is
  kept with the account rather than in any site, because it crosses
  organisations and must never be published with somebody's content; a test
  enforces that nothing which writes into a site can read one.
- An email address can be changed, once the link sent to the new address is
  followed. Until then the old one keeps working, so a typo cannot lock
  anybody out.
- You can see where you are signed in and end one of those sessions, for the
  laptop left at the office.
- An organisation's own owner manages who belongs to it: inviting, changing
  what somebody holds there, and taking them out, rather than asking whoever
  runs the service every time. An invitation that was never accepted can be
  sent again, which stops the previous link, or withdrawn. The people screen
  shows both levels and when each person was last here.
- There is a way out. Every screen now has a person menu with your account
  and signing out, rather than a link on one screen and a picture that did
  nothing on the others; from inside a site there was no way out at all.
  Signing out everywhere ends every session the account has, including the
  one asking.
- An account can be turned off when somebody leaves. They can no longer sign
  in, and whatever they had open stops working at once rather than at its own
  expiry. What they wrote stays attributed to them: history that renames its
  author is history nobody can trust. Only whoever administers the
  installation may do it, because an account is held by the installation and
  somebody in two organisations would otherwise be locked out of one by the
  other.
- Your own account: change your name and your password. The password you have
  is asked for even while you are signed in, so a screen left open on
  somebody else's machine is not enough to take your account over.
- Project settings. What setting a project up decided once can be changed
  afterwards: what it is called, where its site lives, and how it is
  published, using the same panels the setup screen uses rather than copies
  of them. A project can be removed, with a confirmation that says plainly
  that the site itself is not touched.
- The count on the Publish button is the real one. It was never filled in by
  the server and the screens showed a fixed number from the scaffold, so it
  could disagree with the publish screen beside it. Both now read one answer.
- Organisations are real. The address says which one you are looking at, the
  switcher moves between them, and an organisation screen shows its name, its
  address and who is in it. Renaming one keeps its address, so links already
  sent go on working. Until now every screen was told it was in one
  particular organisation, so a second one would have shown the first one's
  work.
- Invitations and password resets are sent. Until now they were issued and
  written to the log, so nobody could be invited without an implementer
  pasting a link by hand. Sending goes through the job queue, so a slow mail
  server delays an invitation rather than losing it. With no mail server
  configured the link still goes to the log, so a small installation does not
  have to set up SMTP before anybody can sign in. Development runs Mailpit,
  started with everything else.
- The two screens those links need: choosing a password from an invitation or
  a reset, and asking for a reset. Setting a password now signs you in, rather
  than asking you to prove twice what you just proved once.
- An application administrator: the person who runs the installation. They
  create organisations and say who owns them, appoint other administrators,
  and see what has been done. Administering is deliberately not editing: it
  grants nothing over any customer's content, so reaching a site means
  joining its organisation, where its owner can see you. The first one is
  named in the configuration and appointed when the server starts.
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
