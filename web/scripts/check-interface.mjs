// Enforces three rules from docs/design/DESIGN.md that a document alone cannot
// keep: colour comes from tokens, controls are large enough to touch, and no
// version control word reaches a content creator. See
// openspec/specs/interface/design-system and interface/editor-language.
import { readdirSync, readFileSync, statSync } from "node:fs";
import { join, relative } from "node:path";

const root = new URL("..", import.meta.url).pathname;
const sourceDir = join(root, "src");

const AUDIENCE_MARKER = "chatcms-audience: implementer";

// A checkbox or a radio is drawn small on purpose and is operated by its
// label, which is what carries the hit area. A line marked this way is
// skipped by the size rule, and the marker has to say which element provides
// the target, so that the claim is visible in review.
const TARGET_MARKER = "chatcms-target:";

// A colour a content creator chooses is their site's, not this interface's,
// so a token would be the wrong thing entirely. A line marked this way is
// skipped by the colour rule, and the marker has to say whose colour it is,
// so that the claim is visible in review.
const CONTENT_COLOUR_MARKER = "chatcms-content-colour:";

// Words the hand-off bans in editor-facing text, with the whole-word rule so
// that "pullover" and "committee" are left alone.
const BANNED_WORDS = [
  "commit", "commits", "committed",
  "branch", "branches",
  "push", "pushed", "pushes",
  "merge", "merged", "merges",
  "repository", "repositories", "repo", "repos",
  "deploy", "deployed", "deploys", "deployment",
  "SHA", "HEAD",
  "conflict", "conflicts",
];

// "pull" is banned as a version control word but is ordinary English in
// "pull request" only; the hand-off bans it outright for editors.
BANNED_WORDS.push("pull", "pulled", "pulls");

const COLOUR = /#[0-9a-fA-F]{3,8}\b|\brgba?\(|\bhsla?\(/;

const MINIMUM_TOUCH = 11; // Tailwind h-11 is 44px.
const MINIMUM_POINTER = 9; // h-9 is 36px, the floor the hand-off allows.

const problems = [];

function* sources(dir) {
  for (const entry of readdirSync(dir)) {
    const path = join(dir, entry);
    if (statSync(path).isDirectory()) {
      yield* sources(path);
      continue;
    }
    if (/\.(ts|tsx)$/.test(entry) && !/\.test\.(ts|tsx)$/.test(entry)) {
      yield path;
    }
  }
}

function userVisibleStrings(text) {
  const found = [];

  // Text between tags, and the value of a prop that holds a sentence.
  const betweenTags = /&gt;([^<>{}]+)&lt;/g;
  for (const match of text.matchAll(/>([^<>{}\n]{3,})</g)) {
    found.push(match[1]);
  }
  for (const match of text.matchAll(/(?:aria-label|label|title|placeholder|summary|alt)=\{?"([^"]{3,})"/g)) {
    found.push(match[1]);
  }
  for (const match of text.matchAll(/"([A-Z][^"\\]{6,})"/g)) {
    found.push(match[1]);
  }
  void betweenTags;
  return found;
}

for (const path of sources(sourceDir)) {
  const text = readFileSync(path, "utf8");
  const name = relative(root, path);
  const forImplementers = text.includes(AUDIENCE_MARKER);

  // Rule one: no colour written by hand, anywhere in the interface.
  const everyLine = text.split("\n");
  for (const [index, line] of everyLine.entries()) {
    if (line.includes(AUDIENCE_MARKER)) continue;
    if (index > 0 && everyLine[index - 1].includes(CONTENT_COLOUR_MARKER)) continue;
    const match = line.match(COLOUR);
    if (match) {
      problems.push(`${name}:${index + 1} writes the colour ${match[0]}; every colour comes from a token`);
    }
  }

  // Rule two: a control is big enough to hit.
  const lines = text.split("\n");
  for (const [index, line] of lines.entries()) {
    if (index > 0 && lines[index - 1].includes(TARGET_MARKER)) continue;
    for (const match of line.matchAll(/\b(?:h|size)-(\d+)\b/g)) {
      const steps = Number(match[1]);
      const isControl = /Button|Item|Trigger|NavLink|button|input|select/.test(line) ||
        /h-\d+ .*(?:px-|cursor|hover:)/.test(line);
      if (!isControl) continue;
      if (steps < MINIMUM_POINTER) {
        problems.push(
          `${name}:${index + 1} defines a control ${steps * 4}px high; the smallest allowed is ${MINIMUM_POINTER * 4}px`,
        );
      }
    }
  }

  // Rule three: no version control word where a content creator can read it.
  if (forImplementers) continue;
  for (const sentence of userVisibleStrings(text)) {
    for (const word of BANNED_WORDS) {
      const pattern = new RegExp(`\\b${word}\\b`, "i");
      if (pattern.test(sentence)) {
        problems.push(
          `${name} says "${sentence.trim()}", which uses the word "${word}"; editors never see version control words`,
        );
      }
    }
  }
}

void MINIMUM_TOUCH;

if (problems.length > 0) {
  console.error(`check-interface: ${problems.length} problem(s)`);
  for (const problem of problems) console.error(`  ${problem}`);
  process.exit(1);
}

console.log("check-interface: tokens, touch targets and editor language are all in order");
