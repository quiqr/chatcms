# The Quiqr model

The model is the declarative description of a site's editable content: which
pages exist, which collections exist, and which fields each of them has. Every
form ChatCMS renders comes from it.

This document is written from the Quiqr desktop app at `../quiqr-desktop`,
which is the behaviour ChatCMS has to reproduce. Where a rule looks arbitrary,
it is, and the source is named so the next reader does not have to guess.

| Behaviour              | Where it comes from                                                             |
| ---------------------- | ------------------------------------------------------------------------------- |
| File locations, assembly, merging | `packages/backend/src/services/workspace/workspace-config-provider.ts` |
| Field types and their properties  | `packages/types/src/schemas/fields.ts`                      |
| Which field types exist           | `packages/frontend/src/components/SukohForm/FieldRegistry.ts` |
| Root properties                   | `packages/types/src/schemas/config.ts`                      |
| Supported formats                 | `packages/backend/src/utils/format-provider-resolver.ts`    |

## Files

A model is not one file.

```
<site>/
  quiqr/
    model/
      base.yaml                  the base file
      includes/
        collections.yaml         merged under the key "collections"
        menu.yaml                merged under the key "menu"
        anything.yaml            merged under the key "anything"
        singles/*.yaml           one single per file
        collections/*.yaml       one collection per file
        menus/*.yaml             one menu section per file
      partials/
        <name>.yaml              pulled in by _mergePartial: <name>
      partialsRemoteCache/       fetched copies of http(s) partials
  sukoh.yaml                     the legacy location of the base file
```

Four formats are read, chosen by extension: `json`, `toml`, `yaml`, `yml`.
`yml` is an alias for `yaml`. A site may mix them freely. When more than one
extension exists for the same name, the first in the order `json`, `toml`,
`yaml`, `yml` wins, because the desktop app globs `{json,toml,yaml,yml}` and
takes the first sorted match.

The base file is `quiqr/model/base.<ext>`. If there is none, `sukoh.<ext>` at
the site root is used. If there is neither, the site has no Quiqr model, which
is a different condition from a model that failed to parse.

## Assembly order

1. Read the base file. Ensure `menu`, `collections`, `singles` and `dynamics`
   exist as lists, creating them empty if absent.
2. Merge every file directly under `includes/`, in filename order.
3. Append every file under `includes/singles/`, then `includes/collections/`,
   then `includes/menus/`, each file contributing exactly one item.
4. Merge each single's, collection's and dynamic's `_mergePartial`.
5. Drop `menu`, `collections`, `singles` and `dynamics` again if they ended up
   empty.

The desktop app has a step between 3 and 4 that merges its own bundled
`dog_food_model`, adding singles for editing Quiqr's own menu and dashboard
from inside the CMS. ChatCMS leaves that out: it is a feature of a desktop
application, not part of a site. A site relying on it sees those singles
reported as unresolved rather than silently missing.

## Merging

The desktop app merges with the JavaScript `deepmerge` package at its defaults:

| Inputs          | Result                                   |
| --------------- | ---------------------------------------- |
| two maps        | merged key by key, recursively            |
| two arrays      | concatenated, the target's entries first  |
| anything else   | the source value wins                     |

Two call sites read backwards at first glance, and both matter.

**An include is merged as `deepmerge(include, base)`.** The base file is the
source, so the **base file wins** on a conflicting key, and an array in the
include comes **before** the base file's entries. The exception is a file whose
name is one of `collections`, `singles`, `dynamics` or `menu` and whose content
is an array: that is concatenated onto the existing list with the base file's
entries **first**.

**A partial is merged as `deepmerge(partial, item)`.** The item is the source,
so the **item wins**, and the partial's array entries come **first**.

**Fields are deduplicated after a partial merge.** The merged `fields` array is
reversed, the first occurrence of each key is kept, and the result is reversed
again. The effect: the **last** declaration of a key survives, which is the
item's, and the surviving fields are ordered by where those last occurrences
fell. A field only the partial declares keeps its place; a field the item
redeclares moves to the item's position. This is cosmetic and it is still
reproduced exactly, because it decides the order of a customer's form.

## `_mergePartial`

A single, collection or dynamic may carry `_mergePartial` instead of repeating
a shared definition. The value resolves as:

| Value                       | Resolves to                                                  |
| --------------------------- | ------------------------------------------------------------ |
| `<name>`                    | `quiqr/model/partials/<name>.<ext>`                           |
| `file://<path>`             | a copy cached under `quiqr/model/partialsRemoteCache`         |
| `http://` or `https://`     | fetched into `quiqr/model/partialsRemoteCache`                |
| `dogfood_site://<name>`     | the desktop app's own bundled partials                        |

ChatCMS reads the remote cache but does not fetch: making the parser reach the
network would turn opening a site into a network operation. An uncached remote
partial is an error naming the partial. A `dogfood_site://` partial is a
warning, since those belong to the desktop app.

After a successful merge the `_mergePartial` key is removed.

## Root properties

| Key           | Meaning                                                          |
| ------------- | ---------------------------------------------------------------- |
| `ssgType`     | Static site generator, default `hugo`                             |
| `ssgVersion`  | Generator version                                                 |
| `hugover`     | The old spelling of the version; migrated to the two keys above   |
| `singles`     | Individual pages, each editing one file                           |
| `collections` | Folders of repeating items                                        |
| `dynamics`    | Carried through, permissively typed                               |
| `menu`        | The sidebar structure                                             |
| `build`       | Named build configurations                                        |
| `serve`       | Named preview-server configurations                               |

### A single

| Key                    | Meaning                                              |
| ---------------------- | ---------------------------------------------------- |
| `key`                  | Identifier, unique among singles                     |
| `title`                | Label                                                |
| `file`                 | The content file this page edits                     |
| `dataformat`           | Front matter format for this file                    |
| `previewUrl`           | Where this page lives on the built site              |
| `description`          | Longer explanation shown in the interface            |
| `pullOuterRootKey`     | Edit a nested key of the file as if it were the root |
| `hidePreviewIcon`      | Hide the preview affordance                          |
| `hideExternalEditIcon` | Hide the external-editor affordance                  |
| `hideSaveButton`       | Hide the save affordance                             |
| `fields`               | The form                                             |
| `build_actions`        | Build actions offered on this page                   |
| `_mergePartial`        | Pull the rest of this definition from a partial      |

### A collection

| Key                | Meaning                                              |
| ------------------ | ---------------------------------------------------- |
| `key`              | Identifier, unique among collections                 |
| `title`            | Label                                                |
| `folder`           | The folder whose files are the items                 |
| `extension`        | File extension of the items                          |
| `dataformat`       | Front matter format of the items                     |
| `itemtitle`        | Singular name of an item, used in labels             |
| `sortkey`          | Field the list is ordered by                         |
| `hideIndex`        | Hide the folder's index file from the list           |
| `previewUrlBase`   | Where items live on the built site                   |
| `hidePreviewIcon`  | Hide the preview affordance                          |
| `fields`           | The form                                             |
| `build_actions`    | Build actions offered on an item                     |
| `_mergePartial`    | Pull the rest of this definition from a partial      |

### The menu

A list of sections. Each section has a `title`, an optional `key`, an optional
`matchRole`, and `menuItems`, a list of objects carrying the `key` of a single
or collection.

## Fields

Every field has these, whatever its type:

| Key                       | Meaning                                                   |
| ------------------------- | --------------------------------------------------------- |
| `key`                     | Identifier, and the key written to the content file        |
| `type`                    | One of the types below                                     |
| `title`                   | Label                                                      |
| `tip`                     | Help text, Markdown                                        |
| `default`                 | Value used when the key is absent                          |
| `hidden`                  | Keep out of the form                                       |
| `name`                    | Alternative identifier used by some container types        |
| `group`                   | Output grouping key                                        |
| `groupdata`               | Group this field's data under its parent                   |
| `arrayTitle`              | Use this field's value as the title of its array item      |
| `content`                 | Markdown body, for layout types                            |
| `theme`                   | Visual theme, for layout types                             |
| `field_prompt_templates`  | Assistant prompt templates, a later milestone              |
| `autoSave`                | Save as soon as the value changes                          |

### Data fields

| Type                | Own properties                                                                  |
| ------------------- | ------------------------------------------------------------------------------- |
| `string`            | `multiLine`, `txtInsertButtons`                                                  |
| `markdown`          | none                                                                             |
| `easymde`           | `multiLine`                                                                      |
| `readonly`          | `multiLine`                                                                      |
| `uniq`              | `multiLine`                                                                      |
| `number`            | `min`, `max`, `step`                                                             |
| `slider`            | `min` (required), `max` (required), `step`                                       |
| `boolean`           | none                                                                             |
| `date`              | `dateFormat`, `date`, `time`                                                     |
| `select`            | `options` (required), `multiple`, `option_image_path`, `option_image_width`, `option_image_extension` |
| `select-from-query` | `query_glob` (required), `query_string` (required), `multiple`, the three `option_image_*` keys |
| `chips`             | `options`                                                                        |
| `color`             | none                                                                             |
| `font-picker`       | `limit`, `families`, `categories`, `variants`                                    |
| `fonticon-picker`   | `multiple`                                                                       |
| `image-select`      | `path` (required), `buttonTitle`, `extensions`, `forceFileName`, `real_fs_path`  |
| `eisenhouwer`       | `dataSetsPath`, `dataSetsKeyToLabel`, `dataSetsLabelPath`, `dataSetsDataPointsPath`, `dataSetsDataPointsKeyToItem`, `dataSetsDataPointPosXPath`, `dataSetsDataPointPosYPath`, `dataSetsDataPointLabelTemplate`, `xScaleTitle`, `yScaleTitle`, `pointRadius`, `labelDoNow`, `labelToPlan`, `labelDelegate`, `labelDelete` |

An option of a `select` is a string, a number, or an object with `value` and
`text`.

### Container fields

| Type             | Own properties                                                              |
| ---------------- | --------------------------------------------------------------------------- |
| `section`        | `fields` (required), `groupdata`                                             |
| `nest`           | `fields` (required), `groupdata`, `max_depth`                                |
| `accordion`      | `fields` (required), `dynFormSearchKey`, `dynFormObjectFile`, `dynFormObjectRoot`, `arrayIndicesAreKeys`, `disableCreate`, `disableDelete`, `disableSort`, `lazy`, `lazyTemp` |
| `pull`           | `fields` (required), `group`                                                 |
| `leaf-array`     | `field` (one child field, not a list)                                        |
| `bundle-manager` | `path` (required), `fields`, `extensions`, `forceFileName`, `maxItems`, `addButtonLocationTop` |

`leaf-array` is the odd one: it carries a single `field`, not a `fields` list,
because every item in the array has the same shape.

### Layout fields

| Type                     | Own properties                                   |
| ------------------------ | ------------------------------------------------ |
| `hidden`                 | none beyond `default`                            |
| `empty-line`             | `amount`                                         |
| `info`                   | `content` (required), `size`, `lineHeight`, `theme` |
| `bundle-image-thumbnail` | `src`                                            |

`info`'s `theme` is one of `default`, `bare`, `warn`, `warn-bare`, `black`,
`black-bare`, `gray`, `gray-bare`.

### An unrecognised type

The desktop app falls back to a `not-found` field rather than dropping the
field or failing the model. ChatCMS does the same: the field is kept, carrying
the type it declared and its raw properties, so a model written for a newer
Quiqr still opens.
