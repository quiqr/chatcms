## Context

`web/` currently holds a placeholder that calls `/api/health`. The design
hand-off in `docs/design/` is the source of truth for everything this change
builds: `DESIGN.md` for the system, `tokens/tokens.css` for the values,
`reference/html/` for exact spacing, and `reference/screenshots/` for the
result. See `proposal.md` for why this comes before the screens.

## Goals / Non-Goals

**Goals:**

- One system, enforced. A rule that is only written down is a rule that
  decays.
- A phone is a first-class target, not a narrowed desktop.
- The editor layout can gain a third panel later without being rebuilt.

**Non-Goals:**

- The screens. Each belongs to the epic that owns its behaviour.
- Dark mode. The hand-off says light only for this milestone, and the tokens
  are named so a dark block can be added later.
- Pixel comparison in the gate. See the assumption in `proposal.md`.

## Decisions

### Rules are checks, not conventions

Three rules from the hand-off are enforced by a script in the frontend check:
no colour written by hand, no control below its minimum size, no banned word
in an editor-facing string. Each is proven to fail before it is trusted.

The alternative is a line in a document and a hope that review catches it. Ten
screens and several epics from now, it will not.

### The banned-word check needs to know who is reading

The hand-off bans git words for content creators and allows them for owners
and implementers in clearly secondary places. A check that cannot tell the two
apart is either useless or unusable.

A file states which it is, on a line of its own:

```ts
// chatcms-audience: implementer
```

A file with no such line is treated as readable by a content creator, so the
safe case is the default and the exception is visible in review. It is a line
in the file rather than a list in a configuration file somewhere else, because
the person adding a technical string is the person who should have to say so.

This is the one place in this project where a comment carries meaning, and it
is specified in `openspec/specs/interface/editor-language` rather than left as
a convention.

### The typefaces are installed, not linked

`tokens/tokens.css` opens with an import from Google Fonts. Keeping it would
mean every reader of every customer's CMS contacts a third party, which for
Dutch customers is a question nobody wants to answer, and it would mean the
built application cannot be checked without a network. The same two typefaces
are installed as packages and served by the application.

Nothing about the design changes: the faces, weights and fallback stacks are
the ones the hand-off names.

### Both frames are rendered; the window picks one

The phone frame and the project frame are both in the document, and which one
shows is decided by the stylesheet. No measuring, no listener, nothing to flash
on first paint, and both are reachable in a test without a browser.

The wider reflow follows the hand-off literally: the shell wraps, and the
content is given a very high flex grow with a large basis, so the navigation
yields first. That is how `reference/html/content.html` does it, and copying
it means the reflow point matches the design rather than a breakpoint someone
chose.

### Placeholder data is marked as placeholder

Screens need something to show before their endpoints exist. The sample
content from the designs lives in `web/src/demo/`, in one place, so that
replacing it with real data later is deleting a directory rather than hunting
through components. The names in it are the designs' invented ones.

### Routing is addresses, not a framework

`react-router` in its declarative form, with one route per milestone screen.
No loaders, no actions, no data layer: those decisions belong with the epics
that fetch real data, and committing to them now would be guessing.

## Risks / Trade-offs

- **A rule that is checked by a script can be worked around.** A colour can be
  built from a string. Mitigated by the rules being cheap to pass honestly:
  every colour the designs use has a token.
- **The audience marking can be forgotten.** Mitigated by the default being
  the strict one: an unmarked file is editor-facing, so forgetting makes the
  check stricter rather than looser.
- **Vendored components drift from upstream.** That is how shadcn is meant to
  work, and the alternative is a dependency that restyles itself on upgrade.
- **Without pixel comparison, a visual difference can land unnoticed.**
  Mitigated by looking at the result against the reference screenshots by hand
  as part of this change, and recording what differs rather than leaving it
  unsaid; the end-to-end epic makes it automatic.

## Migration Plan

The placeholder page is replaced. Nothing is deployed, so there is nothing to
migrate.

## Open Questions

- Whether the sync indicator's implementer tooltip belongs in the shell or in
  the sync epic. It is a slot in the shell here, filled there.
