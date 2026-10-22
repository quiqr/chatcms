# Design

Epic: `.beans/chatcms-2c0p--declarative-workflows.md`

## The file

```yaml
# quiqr/workflows/publish.yaml
on: publish
steps:
  - name: Checking pages and links
    action: check
    run: ./scripts/check-links.sh
  - action: build
  - action: publish
```

`on` is save, preview or publish. `action` is `build`, `check` or `publish`.
`run` belongs to `check` and to nothing else. `name` is what a person is
shown; without one the action supplies its own wording, which is the
designs' own: "Building the site", "Checking pages and links", "Putting it
live".

Unknown keys are refused. A workflow is a thing somebody writes by hand once
and then forgets about, which is exactly when a silently ignored `step:`
instead of `steps:` costs an afternoon.

## Why the order is constrained

Publishing last, at most once; building at most once. A workflow that pushes
and then builds has either a typo or a misunderstanding in it, and neither is
served by running it. The error says so in those words.

## The toolbox

A check runs `sh -c <run>` in the sandbox, against the site's files, with a
fixed set of tools on the path. The set is a setting, supplied by Nix, and
pinned like the git and Hugo binaries already are. A site cannot add to it.

The alternative is letting a workflow name its dependencies, which means
fetching software from the internet in the middle of a publish, on the one
code path whose whole point is that it is contained. The briefing puts
shareable actions as Nix flakes out of scope for milestone 1, and this is
the same question asked early.

## What is sandboxed and what is not

| Step | Runs site code | Sandboxed |
| ---- | -------------- | --------- |
| `check` | Yes, the command the site wrote | Yes |
| `build` | Yes, the site's own templates and content | Yes |
| `publish` | No | No, and it needs the network and the deploy key |

Saying "steps run in the sandbox" without this table would be a comfortable
untruth. The sandbox bounds what the site's own code can do. Publishing is
ChatCMS doing what the project settings say, with no site input to contain,
and it could not be sandboxed anyway: withholding the network and the key is
the sandbox's entire purpose.

## The seam to publishing

The engine does not know how a project publishes. It calls a publisher with
the project and, when a build ran, what was built. Whether that becomes a
push of source to the main line or built pages to a branch of their own is
the publish epic's business, and it is already an epic.

This is also why `build` passing its output to `publish` is a stated
requirement rather than an implementation detail: it is the one thing the two
epics have to agree about.

## What a failure carries

Two things, kept apart. A sentence for the person who pressed the button,
with no git word in it, saying that nothing changed on the live site and the
edits are safe. And the raw output, bounded, which this package never decides
who may see: `policy.Can` with `see-technical-detail` does, where the answer
is being rendered.

Keeping them apart in the type is what makes the second half hard to leak by
accident. A single `Error string` would end up on a screen.

## What this change does not do

No HTTP path fires a trigger yet. Publishing is the epic that gives the
publish trigger a button, and the preview epic the preview one. This change
is the engine and the format, with the defaults that make a freshly imported
site publish correctly before anybody has written a workflow at all.
