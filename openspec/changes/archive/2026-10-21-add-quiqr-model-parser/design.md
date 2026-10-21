## Context

The Quiqr desktop app assembles a site's model in `WorkspaceConfigProvider`
(`../quiqr-desktop/packages/backend/src/services/workspace/`), and types its
fields in `packages/types/src/schemas/fields.ts`. Those two files are the
behaviour this change has to reproduce. See `proposal.md` for why.

The constraint that shapes everything: backwards compatibility is not a
best-effort goal here, it is the product. A model that assembles differently
from the desktop app produces different forms, and different forms write
different files into a customer's repository.

## Goals / Non-Goals

**Goals:**

- Assemble a model byte-for-byte equivalently to the desktop app, including the
  parts of its merge behaviour that are accidents of the library it used.
- Be readable without a site on disk, so the parser can later be handed a git
  working copy without changing.
- Fail loudly and precisely, rather than quietly producing a smaller model.

**Non-Goals:**

- Writing models back. Reading preserves enough to write later, but no writer
  is built here.
- Rendering anything. No field renderer, no form, no editor.
- Evaluating anything the model refers to: no `select-from-query` queries are
  run, no images are resolved, no build actions are executed.

## Decisions

### Three layers, and the merge layer knows nothing about Quiqr

```
format providers     bytes + extension -> generic tree (map[string]any)
assembler            generic trees     -> one generic tree  (merge rules)
schema               generic tree      -> typed Model       (meaning)
```

The assembler is where the compatibility risk lives, and it is the layer that
can be tested exhaustively without knowing what a collection is. Its tests are
tables of input trees and expected output trees. The schema layer then has one
job, turning a known-good tree into Go types, and its failures are
type-and-naming failures rather than silent data loss.

Alternative considered: decoding straight into typed structs per file and
merging typed values. Rejected because merging typed values cannot express
"preserve a key I do not know", which is a requirement.

### Everything is read through `fs.FS`

The entry point takes an `fs.FS` rooted at the site, not a path. Tests use
`fstest.MapFS` and need no temporary directories. `internal/gitsync` can later
hand over a working copy as an `fs.FS` with no change here. It also makes the
"assembly writes nothing" requirement structural rather than a promise.

### The JavaScript merge semantics are reproduced deliberately

The desktop app merges with the `deepmerge` package at its defaults:

| Case                  | Result                                       |
| --------------------- | -------------------------------------------- |
| two maps              | merged key by key, recursively                |
| two arrays            | concatenated, target first                    |
| anything else         | the source value wins                         |

Two call sites matter, and both read backwards at first glance:

- An object include is merged as `deepmerge(include, base)`, so the **base file
  wins** and the include's array entries come **first**.
- A partial is merged as `deepmerge(partial, item)`, so the **item wins** and
  the partial's array entries come **first**.

After a partial merge the fields are deduplicated by key, keeping the **last**
occurrence of each key, and the surviving fields are ordered by where those
last occurrences fall. The effect is that fields only the partial declares keep
their place, and a field the item redeclares moves to the item's position. This
is reproduced exactly, with a test naming it, because a cosmetic difference
here reorders a customer's form.

### Assembly order is fixed and documented

1. Read the base file, and ensure `menu`, `collections`, `singles` and
   `dynamics` exist as lists.
2. Merge `quiqr/model/includes/*` in filename order.
3. Append each file under `includes/singles/`, then `includes/collections/`,
   then `includes/menus/`.
4. Merge each single's, collection's and dynamic's `_mergePartial`.

The desktop app has a fifth step between 3 and 4: merging its own bundled
`dog_food_model`, which adds singles for editing Quiqr's own configuration from
inside the CMS. That is an application feature of a desktop app, not part of a
site, and it is left out. `proposal.md` records what a site relying on it will
see.

### A field is common properties plus a typed configuration

```go
type Field struct {
    Key, Type, Title, Tip string
    Default               any
    Hidden, GroupData     bool
    Fields                []Field       // container types
    Config                FieldConfig   // nil when the type has no own properties
    Extra                 map[string]any
}
```

`FieldConfig` is an interface with one small struct per field type that has
properties of its own: `SelectConfig`, `SliderConfig`, `ImageSelectConfig` and
so on. Types with nothing but the common properties, such as `markdown` and
`uniq`, carry a nil config rather than an empty struct, so an empty config is
never mistaken for a missing one.

Alternative considered: one struct holding the union of every property, which
is how a quick version would do it. Rejected because it makes it impossible to
say which properties a given type actually accepts, and that question is the
whole point of a schema.

### An unrecognised type is a value, not an error

An unknown type becomes a `Field` with `Type` set to what the model declared
and `Config` set to an `UnknownConfig` carrying the raw properties. The desktop
app does the same through its `not-found` field. A model written for a newer
Quiqr therefore opens, and the form epic decides what to draw.

### Errors carry a path, not a sentence

```go
type Error struct {
    File string   // "quiqr/model/includes/collections.yaml"
    Path string   // "collections[2].fields[0].options"
    Want string   // "an array of options"
    Err  error
}
```

Formatting is one place, so every message reads the same way, and the path is
machine-readable for the interface to jump to later.

### The JSON shape is pinned by golden files

The browser will render forms from this JSON, so its shape is a contract from
the moment anything consumes it. Golden files under `test/fixtures/models/`
hold the expected JSON for each fixture model, and a changed shape fails the
gate rather than quietly reaching the frontend.

### Dependencies

| Need | Choice                       | Why                                              |
| ---- | ---------------------------- | ------------------------------------------------ |
| YAML | `github.com/goccy/go-yaml`   | maintained, and reports line and column on error  |
| TOML | `github.com/BurntSushi/toml` | the de facto Go TOML reader, reports positions    |
| JSON | `encoding/json`              | already there; decoded with `UseNumber`           |

Numbers are decoded as `json.Number` and TOML and YAML integers kept as
integers, so a version string like `0.110` in one format does not become a
float in another.

## Risks / Trade-offs

- **The reference behaviour is read from source, not from a specification.**
  Something subtle may be missed. Mitigation: the fixture models are taken from
  the reference app's own test fixtures and from the customer site, and the
  assembled output is compared against what the desktop app produces, not
  against what this change expects.
- **`deepmerge`'s array concatenation can duplicate entries** when a site
  declares the same collection in both a base file and an include. The desktop
  app has the same behaviour. Mitigation: reproduce it, and report duplicate
  keys as a model error so the site author sees it rather than a doubled menu.
- **Excluding the dog food model is a visible behaviour change.** Mitigation:
  stated in the proposal, and the missing singles are reported as model
  warnings naming the `dogfood_site://` partial that could not be resolved.
- **27 field types is a lot of surface for one change.** Mitigation: the
  inventory document is written first, from the reference app, and the test
  table is generated from it, so a type cannot be forgotten silently.

## Open Questions

- Whether the `dynamics` list is used by any real site. It is carried through
  assembly and typed permissively, which is enough either way; if no customer
  site uses it, a later change can drop it from the schema.
