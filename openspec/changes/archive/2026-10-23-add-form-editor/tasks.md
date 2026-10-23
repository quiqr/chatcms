# Tasks

## 1. A page as a form

- [x] 1.1 Locate a page from the model: a single, or an item of a collection
- [x] 1.2 Refuse an item name that could reach outside its collection
- [x] 1.3 Describe the controls, with values, groups and lists
- [x] 1.4 Show a kind of field this version cannot draw, as text, saying so
- [x] 1.5 Leave the page's own text to its own editor, not drawn twice

## 2. Writing an edit back

- [x] 2.1 Write values by path through `internal/forms`
- [x] 2.2 Refuse a field the page does not have
- [x] 2.3 Refuse text where the model gives it no field
- [x] 2.4 Make a version, attributed to whoever edited
- [x] 2.5 Ask for a sync, and do not fail the save when that cannot be queued
- [x] 2.6 Tell a site with a mistake in its model apart from one with no model

## 3. The HTTP layer

- [x] 3.1 Read and write one page
- [x] 3.2 Wire it into the server
- [x] 3.3 Handler tests with a stub service

## 4. The controls

- [x] 4.1 Short text, long text with its length, number, slider
- [x] 4.2 Date, choice, yes or no, colour
- [x] 4.3 A list of values, and a list of items that opens
- [x] 4.4 A group that folds
- [x] 4.5 Read-only, a note, and the one shown as text with a word about it
- [x] 4.6 A marker in the interface check for a colour a person picks

## 5. The text editor

- [x] 5.1 The toolbar from the designs, writing Markdown around a selection
- [x] 5.2 A source view showing exactly what is saved
- [x] 5.3 Shown but not editable where the model gives it no field

## 6. Saving by itself

- [x] 6.1 Hold an edit until typing stops, then write it once
- [x] 6.2 The three states, and no Save button anywhere
- [x] 6.3 Write what is waiting when the screen goes away
- [x] 6.4 A test that no git word reaches what an editor reads

## 7. Shipping

- [x] 7.1 The seed gives a site every kind of field, so this can be looked at
- [x] 7.2 `docs/ARCHITECTURE.md`
- [x] 7.3 The gate green, coverage held
