# The screens next to the drawings

What `docs/design/reference/screenshots/` draws, what the product does, and
what was decided about each difference. Written by reading both, not by
comparing pixels: a drawing from a design tool and a browser rendering a
different font engine differ in a thousand places that mean nothing, and a
comparison that cries wolf is one nobody reads.

Capture the screens again with:

```bash
cd test/e2e
CHATCMS_BIN=... CHATCMS_SEED=... CHATCMS_SHOTS=../../shots \
  npx playwright test --project=shots-desk --project=shots-phone
```

Last read: 23 October 2026, against the fixture site the seed makes.

## Decided as differences to keep

| Where       | The drawing                                     | The product                                        | Why                                                                |
| ----------- | ----------------------------------------------- | -------------------------------------------------- | ------------------------------------------------------------------ |
| Content     | A column of pages and collections beside a list  | One column, pages then collections, each a section  | The second column is worth having once a site has more than a handful of collections. Milestone 02. |
| Content     | Thumbnails on every row                          | An icon                                             | A thumbnail needs a picture per page, which the model does not name. |
| History     | A detail panel beside the list                   | The detail under the entry it belongs to            | Works at both widths without a second layout. |
| Publish     | The changes in a bordered card                   | A plain list                                        | Cosmetic. |
| Sync        | A sheet of every indicator state                 | One screen showing the state the site is in         | The drawing is a style sheet, not a screen; the states it draws are the ones `SyncIndicator` has. |
| Onboarding  | A full-width wizard with numbered steps          | The same steps inside the project shell             | Leaving the shell to set a project up loses the way back. |
| Login       | Two cards illustrating a form and a preview      | The sentence alone                                  | Illustration, not interface. |
| Everywhere  | The screen's name only in the page               | The name in the top bar and again in the page       | The top bar names where you are; the page names what you are reading. Worth revisiting once there is a page whose two names differ. |

## Still to do

| Where       | The difference                                                                | What it needs                                      |
| ----------- | ------------------------------------------------------------------------------ | -------------------------------------------------- |
| Projects    | No "Edited by … · N people" line, and the status is a word rather than a badge  | Both need the project's people and its last edit    |
| Media       | No filter by collection                                                         | Needs where each picture is used, which is read but not yet filtered on |
| Editor      | No overflow menu on a phone                                                     | Nothing in it yet that a phone cannot reach another way |
| Onboarding  | "Push to" and "What to push" are not labelled as groups                         | Two headings |

## Found by the comparison and fixed

- The editor's top bar said `home`, the key of the page, where the drawing
  says which page it is. It now says the collection and the title, and the
  form no longer repeats them.
- A person who has not accepted their invitation had their email address
  printed twice, once where a name would be.
- The sign-in screen had lost its subtitle and its footer.
- The Edit and Preview switch on a phone was half the width of the screen.

## About the captures themselves

- The editor's preview frame is empty in the capture. Hugo is still starting
  when the screen is caught; the acceptance test waits for it properly.
- The phone captures are full-page, so the bar that is fixed to the bottom of
  the screen appears part way down the image. That is the capture, not the
  screen.
- The media screen is empty because the fixture site has no pictures yet.
