# Design

Epic: `.beans/chatcms-sbe1--form-rendering-and-round-trip-save.md`, second of
its two changes.

## What the browser is given

A page comes back as controls with values, not as a model the browser has to
interpret:

```json
{"path": "seo", "type": "section", "title": "Search", "fields": [
  {"path": "seo.description", "type": "string", "title": "Description",
   "value": "A bakery."}
]}
```

The path is what a value is keyed by when it goes back. Working that path out
is the binder's job in `internal/forms`, which already knows that a grouping
field may or may not group its data. Sending the model and letting the browser
derive paths would mean writing those rules twice, in two languages, and
having them drift.

An edit goes back as `{"values": {"seo.description": "..."}}`: whole values,
by path. The file is what keeps order and formatting, so the transport does
not have to.

## A field we cannot draw

A Quiqr model can declare kinds of field this version has no control for. Two
honest options: leave them out, or show them as text.

Leaving them out makes a page look emptier than it is, and the first save
would then be a save of a form that does not know about them. Showing them as
text, saying so, means nothing is invisible and nothing is lost. A test walks
a model with a field nobody has implemented and checks it is both shown and
saved.

## The page's own text

The reference implementation edits a page's Markdown through a field with a
particular key. A model that declares none is a model whose pages have no
editable text, and that is a decision the site made, not one to work around.

So the text is shown either way and written only where there is a field for
it. An attempt to write it where there is not is refused with a reason rather
than quietly dropped, because quietly dropping somebody's writing is the worst
thing an editor can do.

That field is also not drawn twice. It is the page's text, so the text editor
draws it; drawing it again as an ordinary field would be two boxes for one
thing, which is exactly what the first run of this looked like.

## Saving by itself

No Save button, as the designs ask. An edit is held for a moment and written
when typing stops, so a sentence is one save rather than forty. What is
waiting is written when the screen goes away, because the last thing typed is
the thing most likely to be lost.

Three states, which are the sync indicator's: saved, saving, and not saved
yet. The third says it will keep trying rather than asking somebody to do
something about it.

## A colour is not a colour

The interface check refuses a hand-written colour, because colour comes from
tokens. A colour picker's fallback is not this interface's colour at all: it
is the site's, chosen by the person. The check gained a marker for that, which
has to name whose colour it is, so the claim is visible in review rather than
a silent exemption.

## What this change does not do

Browsing, creating, renaming and deleting pages is content management, the
next epic, which this one unblocks. The preview column beside the form is
still a placeholder; the preview epic fills it, and the field-to-preview link
belongs there because it needs a preview to point at.
