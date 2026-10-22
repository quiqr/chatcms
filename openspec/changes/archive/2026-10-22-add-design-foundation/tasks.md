## 1. The foundation

- [x] 1.1 Merge the tokens from `docs/design/tokens/tokens.css` into the
  application stylesheet under the shadcn names, without the font import;
  verify with `tokens are all present` that every token the hand-off defines
  exists.
- [x] 1.2 Install Geist and Geist Mono as packages and serve them from the
  application; verify with `no third-party font request` that the built
  output names no font service.
- [x] 1.3 Copy the logo into the application's public files and render it in
  the header at the size the hand-off gives; verify it is served.
- [x] 1.4 Add the utility that merges class names, and the variant helper the
  components use; verify with `merges conflicting classes`.

## 2. The components

- [x] 2.1 Add the button with its primary, ink, outline and ghost variants and
  its sizes; verify with `button renders each variant` and
  `publish button carries a count`.
- [x] 2.2 Add the status badge with the states the designs use; verify with
  `badge renders each state` and that the ones with a dot have one.
- [x] 2.3 Add the segmented control; verify with `segmented control selects`
  that choosing an option reports it and that the selected option is marked
  for assistive technology.
- [x] 2.4 Add radio cards; verify with `radio cards select one` including
  keyboard operation.
- [x] 2.5 Add the menu, the dialog and the disclosure; verify each opens,
  closes and returns focus.
- [x] 2.6 Add the toast, the table wrapper, the tooltip, the avatar and the
  project tile; verify each renders and that the table wrapper scrolls on its
  own rather than widening the page.
- [x] 2.7 Give every icon-only control a readable name; verify with
  `every icon-only control is named`.

## 3. The shells

- [x] 3.1 Build the project shell with its sidebar, top bar, sync slot and
  publish action; verify with `project shell shows navigation` and
  `publish action shows the count`.
- [x] 3.2 Hide the setup sections from editors and viewers rather than
  disabling them; verify with `setup is absent for an editor`.
- [x] 3.3 Make the shell reflow by the hand-off's rule rather than a chosen
  breakpoint; verify with `content yields last` against the class contract.
- [x] 3.4 Build the phone shell with its header and bottom tab bar, padded
  clear of the device; verify with `phone shell has its tabs` and
  `bottom bar clears the device`.
- [x] 3.5 Render both frames and let the stylesheet choose; verify with
  `only one frame is visible at a time` that the other is hidden rather than
  absent.
- [x] 3.6 Build the organisation header with its switcher; verify with
  `organisation header replaces the sidebar`.
- [x] 3.7 Build the editor frame as a row of slots with a narrow rail and an
  empty third slot; verify with `editor has three slots` and
  `editor uses the rail`.
- [x] 3.8 Swap the phone tab bar for the editing actions while editing; verify
  with `editing replaces the tabs`.

## 4. Routes and placeholder data

- [x] 4.1 Add a route for each of the ten milestone screens; verify with
  `every milestone screen has an address`.
- [x] 4.2 Answer an address that names nothing in plain language with a way
  back; verify with `an unknown address is explained`.
- [x] 4.3 Put the designs' sample content in one place, marked as placeholder;
  verify with `placeholder data is in one place`.

## 5. The rules

- [x] 5.1 Add the check that fails on a colour written by hand in a component;
  verify it fails on a deliberate breach and passes on the tree.
- [x] 5.2 Add the check that fails on a control defined below its minimum
  size; verify it fails on a deliberate breach.
- [x] 5.3 Add the banned-word check with the audience marking, defaulting to
  editor-facing; verify it fails on a deliberate breach, allows a marked file,
  and does not fail on a longer ordinary word.
- [x] 5.4 Run all three from the frontend check alongside the type check and
  the tests; verify `nix flake check` runs them.

## 6. Looking at the result

- [x] 6.1 Render the shell at 1440 and at 390 and compare with
  `docs/design/reference/screenshots/desktop/content.png` and
  `phone/phone-content.png`; record every difference in this change rather
  than leaving it unsaid.
- [x] 6.2 Check at 390 that no page scrolls sideways and no touch target is
  under forty-four pixels; record the result.

## 7. The gate and shipping

- [x] 7.1 Update the pinned `npmDepsHash` in `flake.nix`; verify
  `nix build .#web` succeeds.
- [x] 7.2 Add a bullet to `CHANGELOG.md` under `## [Unreleased]`.
- [x] 7.3 Run `nix flake check` and confirm every check passes.
