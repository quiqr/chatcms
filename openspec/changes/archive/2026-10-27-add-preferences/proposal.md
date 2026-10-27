## Why

Nothing about how the application behaves can be chosen. Dates are written
in whatever the browser's locale happens to be, in four places with three
different formats. There is one appearance, with no dark one and no way to
ask for either, though `docs/design/DESIGN.md` section 11 says the tokens
are named so that a dark block can be added.

Epic `chatcms-6h9u`.

Bean: `.beans/chatcms-6h9u--account-preferences-language-dates-and-a-dark-appe.md`.

## What Changes

- **A preference is kept on the account**, so it follows somebody to another
  machine rather than being lost the first time they use one.
- **A dark appearance**, and following what the machine asks for, which is
  the default and is not a choice anybody has to make.
- **One way to write a date and a time**, read by every screen, instead of
  the same decision made four times.
- **A preferences screen**, beside the account rather than inside it: one is
  who you are, the other is how you like things.

## Capabilities

### New Capabilities

- `identity/preferences`: what a person may choose about how the application
  behaves, and where that choice is kept.

### Modified Capabilities

None.

## Impact

- **New**: a migration for the preference, endpoints to read and change it,
  `web/src/routes/PreferencesPage.tsx`, a dark set of tokens, and one place
  that writes a date.
- **Changed**: four screens stop formatting dates themselves.

## What this change does not do

**Dutch.** The bean also asks for the interface in Dutch, and that is not
here. The mechanical half, one place for every string, is a change to 62
files and some 350 pieces of text; the other half is the product's voice in
a language this repository has a style guide about, and generating it
wholesale would produce Dutch that reads like a machine wrote it. It is
`chatcms-31tm`, written so that the extraction and the copy are done
together by somebody who can judge the result.

## Assumptions

- **Following the machine is the default, and not a choice.** Somebody who
  never opens this screen gets what their system asks for. A product that
  depends on people having chosen is a product that looks broken until they
  do.

- **The preference is on the account, not in the browser.** Browser storage
  is lost the first time somebody uses another machine, and this is the kind
  of thing people set once.

- **A dark appearance is a second set of tokens, not a second stylesheet.**
  `web/scripts/check-interface.mjs` already refuses a hand written colour,
  so the only way to add one is to name it, which is what makes this
  possible at all.

- **Dates are written in one place.** Four screens each decided for
  themselves, with three different formats; `docs/design/DESIGN.md` section
  7 asks for relative and human, which is one rule and belongs in one
  function.
