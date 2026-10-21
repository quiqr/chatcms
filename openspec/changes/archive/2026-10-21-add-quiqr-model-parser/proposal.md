## Why

The Quiqr model is the heart of the product: a declarative YAML description of
a site's singles, collections and fields, from which every editing form is
generated. Nothing downstream can be built without reading it, and it must read
today's models unchanged. Epic `chatcms-vho6`.

It serves acceptance test 3 from `docs/BRIEFING-M01.md` section 6 ("all forms
from the model render"), and it is the precondition for acceptance tests 4, 5
and 7.

## What Changes

- **The model is read the way the desktop app reads it.** A site's model is not
  one file. It is a base file, plus include files merged by filename, plus
  per-item include directories, plus partials pulled into individual singles
  and collections. The assembly order decides what wins, and getting it wrong
  silently changes a customer's forms.
- **Three serialisation formats.** Model files may be YAML, TOML or JSON, mixed
  freely within one site. The format is chosen by file extension.
- **A typed schema.** Singles, collections, dynamics, the menu, the build and
  serve entries, and the twenty-seven field types the desktop app supports,
  including the five container types that nest further fields.
- **Nothing is dropped.** Keys this version does not understand are preserved
  rather than discarded, and an unknown field type becomes a placeholder that
  names itself rather than disappearing from the form.
- **Validation that points at the mistake.** An invalid model names the file,
  the path within it, and what was expected.
- **A derived content navigation.** The Pages and Collections lists the content
  browser needs, with item counts and the per-collection item name used in
  labels such as "New article", come from the model rather than from a second
  hand-maintained structure.
- **The model has a stable JSON shape.** The browser will render forms from
  data rather than from hand-written forms per site, so the model serialises to
  a documented JSON representation that is pinned by golden tests.
- **An inventory document.** Every field type the desktop app supports, with
  its properties and semantics, written from the reference app and committed as
  `docs/QUIQR-MODEL.md`, so the form epic has one place to work from.

## Capabilities

### New Capabilities

- `content/model-source`: how a site's model files are found, read and
  assembled into one model, across formats, includes and partials.
- `content/model-schema`: what a model contains once assembled, which field
  types exist, and what is preserved or rejected.
- `content/content-navigation`: the Pages and Collections structure derived
  from a model that the content browser and the editor navigate by.

### Modified Capabilities

None. The parser adds no endpoint; see the assumption about where the model is
served from.

## Impact

- **New code**: `internal/model`.
- **New documentation**: `docs/QUIQR-MODEL.md`.
- **New fixtures**: `test/fixtures/models/`, covering every field type, all
  three formats, includes, sub-includes and partials.
- **Dependencies added**: a YAML parser that preserves key order and a TOML
  parser. Both pinned, both vendored through `vendorHash`.
- **Seams**: none of the four seams in `docs/ARCHITECTURE.md` are touched. The
  parser reads a directory tree given to it and never reaches for git itself,
  so `internal/gitsync` can supply a working copy later without changing it.
- **Coverage**: `internal/model` becomes the first core package to carry the
  80% floor, which the gate has been reporting as absent since the foundation.

## Assumptions

- **The desktop app's own model is not carried over.** The reference app merges
  a bundled `dog_food_model` into every site, which adds singles for editing
  Quiqr's own menu and dashboard from inside the CMS. That is an application
  feature of the desktop app, not part of a site's model, and ChatCMS has its
  own settings screens. It is excluded, and a site that depends on it will show
  those singles as missing rather than silently broken.
- **Remote partials are read from cache only.** `_mergePartial` values pointing
  at `http://` or `https://` are fetched by the desktop app and cached under
  `quiqr/model/partialsRemoteCache`. Fetching over the network from inside the
  parser would make reading a model a network operation. This change reads the
  cache when it is present and reports a clear error when it is not; fetching
  belongs with the sync engine, which already owns the network.
- **`hugover` is migrated, not honoured.** The desktop app migrated the old
  `hugover` key to `ssgType` and `ssgVersion`. This change does the same
  migration on read, and treats the result as the single source of truth.
- **Unknown field types render as a placeholder.** The desktop app falls back to
  a `not-found` field. The same behaviour is kept, so a model written for a
  newer Quiqr still opens.
- **Field types outside milestone 1's editing scope are still parsed.** Types
  such as `eisenhouwer` and `select-from-query` are read, carried and served, so
  a model containing them is not rejected. Whether the browser can edit them is
  the form epic's decision, not the parser's.
- **No HTTP endpoint yet.** Serving a model means naming which site's model,
  and a site is a project, which the tenancy and onboarding epics own. Adding a
  stand-in addressing scheme now would be a crutch to remove later. This change
  delivers the parser and its JSON shape; the content management epic adds the
  endpoint once projects exist.
