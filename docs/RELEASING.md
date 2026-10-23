# Releasing ChatCMS

A release is a version number, a changelog entry and a tag. Everything else
is already automated.

## The version

`version` in `flake.nix` is the one place it is written. The build stamps it
into the binary:

```
-X=github.com/quiqr/chatcms/internal/version.Version=${version}
```

`chatcmsd --version` and `/api/health` both report it, so a running machine
can be asked what it is rather than inferred from a deployment log.

## Steps

1. **The gate.** `nix flake check`. A release is never made from a tree the
   gate has not passed.
2. **Promote the changelog.** Move everything under "Unreleased" in
   `CHANGELOG.md` into a section headed with the new version and today's date.
   Leave "Unreleased" empty rather than removing it.
3. **Set the version** in `flake.nix`.
4. **Commit** both: `jj commit -m "Release 0.2.0"`.
5. **Tag** the commit `v0.2.0`.
6. **Push** the branch and the tag.

## What goes in the changelog

What changed for somebody using it, not what changed in the code. An entry
nobody outside this repository could act on belongs in the OpenSpec archive,
which is where the reasoning already is.

Every archived change in `openspec/changes/archive/` is a release note waiting
to be written: its proposal says why, in the words the change was argued in.

## Versions before 1.0

The schema may change in either direction and no promise is made about
upgrading between pre-release versions other than by migration, which runs
automatically. Once a customer is on it, that stops being true, and this
section is rewritten rather than quietly ignored.
