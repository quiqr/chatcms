## Why

The preferences screen says, in as many words, that Dutch is being worked
on. It is not: every string is written where it is shown, in some sixty
files, and there is nowhere to put a second language.

The first customer is Dutch. A product whose argument is that it speaks to
people who are not technical cannot ask them to read English.

Epic `chatcms-31tm`.

Bean: `.beans/chatcms-31tm--dutch-and-one-place-for-every-string.md`.

## What Changes

- **One catalogue**, with every string in it and nothing written where it is
  shown. The English catalogue is the source, and another language is a file
  typed against it, so a missing string is a failed type check rather than a
  word in the wrong language on somebody's screen.
- **A check that keeps it that way**, failing on a string written in a
  screen the way the interface check already fails on a colour written by
  hand.
- **Dutch**, written to the same copy rules the English follows.
- **A language on the account**, beside the appearance, falling back to what
  the browser asks for. It follows to another machine, and takes effect at
  once.
- **Dates and times follow it**, through the one function that already
  writes them.
- **A refusal from the server is said in the catalogue too.** Each one
  carries a key; the screen says what the catalogue has for that key, and
  falls back to what the server said.

## Capabilities

### New Capabilities

- `interface/language`: one place for every string, how another language is
  added, and what the check holds.

### Modified Capabilities

- `identity/preferences`: a chosen language, kept with the account.
- `interface/editor-language`: the banned words are checked in the
  catalogue, in both languages.

## Impact

- **New**: `web/src/strings/`, `web/src/lib/say.ts`, a check script, a
  column on a person, and the language on the answer every screen already
  asks for.
- **Changed**: every screen reads its text from the catalogue;
  `httpapi.RefusedError` carries a key; `web/src/lib/when.ts` writes dates
  in the chosen language.

## Assumptions

- **The English catalogue is the type.** Another language is
  `Record<keyof typeof en, string>`, so adding Dutch made every missing
  string a compile error, and adding a third language will do the same.
  There is no runtime report of what is missing, because a missing string is
  not allowed to reach a build.

- **The language is applied at the module, and the screen is redrawn.**
  `say` is a plain function rather than a hook, so that text can be written
  where there is no component: a helper, a date, an error. Choosing a
  language redraws everything below, which is a moment somebody has just
  asked for.

- **A key reads like what it says.** Keys are grouped by the screen they
  belong to, and the English is in the key's own file beside it, so that a
  translator reads them in order rather than hunting.

- **The server's refusals are said by the screen.** The server keeps its
  English reason, which is what a log and a developer want, and adds a key.
  The screen says the catalogue's version when it knows the key. A refusal
  nobody has translated yet therefore appears in English rather than not at
  all.

- **Dutch keeps the accepted name of a technical thing.** The Dutch writing
  guide treats an English word as a red flag; this interface is for people
  who are not technical, so almost nothing here is a technical term at all.
  Where one is unavoidable in text for implementers, the accepted name
  stays.

- **The acceptance suite runs once in Dutch.** Not every spec twice: one
  pass that signs in, chooses Dutch, and walks the screens a content creator
  uses, which is what catches a string nobody moved.
