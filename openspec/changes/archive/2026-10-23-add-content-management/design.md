# Design

Epic: `.beans/chatcms-tp3b--content-management.md`

## A name is a file name

The old Quiqr names a page's file after its title. That is why renaming is a
move rather than a field change, and why it also rewrites the title inside the
page: to a content creator those are the same act, and leaving them apart
would mean a page called one thing and addressed as another.

The folding is deliberately small: the letters a Dutch or German site actually
uses, folded to what they are made from, and anything else becomes a hyphen.
That is what the reference does, and a bigger table would be a different
product's behaviour.

A title that would become a path is folded rather than refused. Somebody
typing "A/B test" means a slash in words, not a folder, and `a-b-test` is what
they meant. A title with no letters or digits at all is refused, because there
would be nothing left to name the file.

## The mark comes from git

Whether an item is waiting to be published is the same question the publish
review asks, answered by the same method. A flag the app set by hand would be
a second source of truth, and the two would disagree the first time anything
went wrong.

## Removing is a save

A removed page goes from the site at the next publish and stays in the
history, which is exactly what going back already does. So removing needs no
second mechanism and the confirmation can honestly say the page can be brought
back.

## Where the New project button went

It had been drawn since the design foundation and wired to nothing. The
backend endpoint shipped with tenancy; the setup screen shipped with
onboarding; nobody's epic owned the step between them, so the only way to get
a project was the development seed.

It belongs here because this epic owns the screens a project is reached
through. The dialog asks for the two things a project needs, suggests the
address from the name, and then goes to the setup screen that already exists.
The editor's Publish button was inert in the same way and is now a link.

## What this change does not do

Adding a page the model does not name means editing the model, which is an
implementer's job and a later epic. Media is its own epic, so a collection's
items have no thumbnails yet.
