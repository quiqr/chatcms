# Tasks

## 1. The harness

- [x] 1.1 A PostgreSQL, a bare remote and a server per spec
- [x] 1.2 Seed a site worth testing against: every field type, a model, pages
- [x] 1.3 Wait for the server to answer, with a deadline and no sleep
- [x] 1.4 Tear everything down, including after a failure
- [x] 1.5 Reach the remote from a spec, to check what really arrived

## 2. The nine specs

- [x] 2.1 An organisation and an invitation
- [x] 2.2 Importing a site with a key
- [x] 2.3 Forms from the model, and saving byte for byte
- [x] 2.4 A page made, edited, removed, and a picture uploaded
- [x] 2.5 The preview follows the form
- [x] 2.6 Work pushed elsewhere arrives by itself
- [x] 2.7 Publishing shows what changed, then sends it
- [x] 2.8 Going back, and going back from going back
- [x] 2.9 The screens no other spec lands on, on a phone

## 3. Both widths

- [x] 3.1 Run every spec at 1440 px and at 390 px
- [x] 3.2 No sideways scrolling, on every phone pass
- [x] 3.3 Nothing under 44 px, on every phone pass
- [x] 3.4 Fix what the phone rules found, in the product

## 4. In the gate

- [x] 4.1 The `e2e` check in `flake.nix`, browser and all pinned
- [x] 4.2 Green three times in a row, with no flake
- [x] 4.3 A deliberately broken behaviour fails its own spec and no other

## 5. Against the drawings and against a real site

- [x] 5.1 Capture every screen at both widths
- [x] 5.2 Write the differences down in `docs/design/COMPARISON.md`
- [x] 5.3 `scripts/acceptance-against.sh`, to run the suite on a real site
