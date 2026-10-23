## Why

The editor can open a page, but only if somebody types its address. A content
creator needs to see what the site holds and to make, rename, copy and remove
pages. Epic `chatcms-tp3b`, and the page half of acceptance test 4.

It also closes a hole nobody's epic owned: the New project button has been
drawn since the design foundation and wired to nothing, so the only way to get
a project has been the development seed.

## What Changes

- **A content browser**: the site's pages and each collection, in the order
  the model asks for, with a mark on anything waiting to be published, search,
  and "show more" for a long collection.
- **A new page** in a collection, named from what somebody typed, with the
  file named the way the old Quiqr names it.
- **Rename, make a copy and remove**, with removing asking first and saying
  what will happen.
- **A page the model names cannot be renamed or removed here.** Its name is
  part of how the site is built.
- **The New project button works**: a name, a suggested address, an optional
  domain, and then the setup screen that already exists.
- **The editor's Publish button goes to the publish screen**, which it did not.

## Capabilities

### New Capabilities

- `content/browsing`: what a content creator sees of their site, and how it is
  ordered and marked.
- `content/pages`: making, renaming, copying and removing a page, and what the
  files end up called.

### Modified Capabilities

None.

## Impact

- **New code**: `internal/content/browse.go`, `internal/content/manage.go`,
  `web/src/routes/ContentPage.tsx`.
- **Changed**: the projects screen can make a project; the editor's Publish
  button is a link; `content.Site` gains the question about what is waiting.
- **Schema**: none.

## Assumptions

- **A file is named after its title, as the old Quiqr names it.** Small
  letters, digits and hyphens, with accents folded. That is why renaming is a
  move rather than a field change, and why renaming also changes the title
  inside the page: the two are one thing to a content creator.

- **A title that would become a path is folded, not refused.** Somebody typing
  "A/B test" means a slash in words. It becomes `a-b-test`. A title with no
  letters or digits at all is refused, because there would be nothing to name
  the file.

- **Removing is a save, not a hard deletion.** The page goes from the site at
  the next publish and stays in the history, which is what the confirmation
  says. Nothing needs a second kind of deletion.

- **Only a collection can grow.** Pages the model names are the site's shape;
  adding one means editing the model, which is an implementer's job and a
  later epic.

- **The browser reads what is waiting from git, not from a flag.** A mark on
  an item comes from the same question the publish review asks, so the two can
  never disagree.
