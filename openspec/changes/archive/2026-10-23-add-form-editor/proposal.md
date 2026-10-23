## Why

The editor is the product. Everything shipped so far gets a site in, keeps it
safe, publishes it and puts it back; none of it lets anybody change a word.
The second half of epic `chatcms-sbe1`, and what content management and live
preview are both waiting behind.

## What Changes

- **A page is read as a form**: the controls to draw, with the values the file
  holds, worked out from the site's own Quiqr model.
- **Every field type the model can declare is drawn**, and a kind this version
  has no control for is still shown, as text, saying so. Nothing a site holds
  is invisible.
- **A page's own text is edited with a Markdown editor**, with the toolbar
  from the designs and a source view, but only where the model gives that text
  a field. Where it does not, the text is shown and left alone.
- **An edit is written back through `internal/forms`**, so the file keeps its
  key order, its quoting and its spacing, and only what changed changes.
- **Saving happens by itself**, a short while after typing stops, with the
  three states the designs draw and no Save button anywhere.
- **Every save makes a version and asks for a sync**, so an edit reaches the
  remote without anybody waiting for the next round.

## Capabilities

### New Capabilities

- `content/editing`: what a page looks like as a form, and what writing one
  back does.
- `interface/form-fields`: what a person sees for each kind of field, and what
  is never lost.

### Modified Capabilities

None.

## Impact

- **New code**: `internal/content`, `web/src/fields/`,
  `web/src/routes/PageEditor.tsx`.
- **Changed**: the editor screen's form column holds the real form;
  `internal/auth` gains a person lookup by identifier.
- **Schema**: none. A page lives in the site.
- **The interface check** gains a marker for a colour a content creator picks,
  which is the site's colour and not this interface's.

## Assumptions

- **A field this version cannot draw is shown as text rather than left out.**
  Leaving it out would make a page look emptier than it is, and saving it
  would then lose whatever was in it. Shown as text, it says plainly that it
  is shown as text, and what somebody types is written exactly as typed.

- **A page's text is editable only where the model says so.** The reference
  implementation edits that text through a field with a particular key, so a
  model that declares none is a model whose pages have no editable text. The
  text is still shown, because hiding it would be worse than not editing it,
  and an attempt to write it is refused rather than quietly dropped.

- **Clearing a field writes the absent value the reference writes.** It does
  not remove the key. That is what `internal/forms` already does, and it is
  part of the backwards-compatibility contract rather than this change's
  decision.

- **Saving is by itself and whole-field.** An edit carries the value of each
  control that changed, not a patch. The file is the thing that keeps order
  and formatting; the transport only has to carry what a person typed.

- **Browsing pages is content management's epic.** This change reads and
  writes one page, addressed by what the model calls it. Listing, creating,
  renaming and deleting are the next epic, which this one unblocks.
