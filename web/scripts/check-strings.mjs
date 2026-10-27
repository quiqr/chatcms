// Holds the rule that no string a person can read is written where it is
// shown: every one of them comes from web/src/strings, by key. See
// openspec/specs/interface/language.
//
// It also reads the catalogue itself, in every language, so that the words
// the design hand-off bans cannot come back in through a translation.
import { readdirSync, readFileSync, statSync } from "node:fs";
import { join, relative } from "node:path";
import ts from "typescript";

const root = new URL("..", import.meta.url).pathname;
const src = join(root, "src");

// A line marked this way is left alone, and the marker has to say why, so
// that the claim is visible in review. A file whose head carries the marker
// is left alone whole, which is for a file that holds no interface text at
// all: the names of the screens, say, which only a test reads.
const MARKER = "chatcms-not-a-string:";

// Where a string is read by somebody rather than by the machine.
const SAYS = new Set([
  "label", "placeholder", "title", "description", "hint", "aria-label",
  "alt", "summary", "heading", "caption", "note",
]);

// Every key the catalogue holds. A prop that carries one of these is a key
// being passed on, not a sentence written where it is shown: a route says
// which title a screen has, and the screen says it.
const keys = new Set(
  [...readFileSync(join(src, "strings/en.ts"), "utf8").matchAll(/^  "([^"]+)":/gm)].map(
    (match) => match[1],
  ),
);

const problems = [];

function* sources(dir) {
  for (const entry of readdirSync(dir)) {
    const path = join(dir, entry);
    if (statSync(path).isDirectory()) {
      yield* sources(path);
      continue;
    }
    if (!/\.(ts|tsx)$/.test(entry)) continue;
    if (/\.test\.(ts|tsx)$/.test(entry)) continue;
    yield path;
  }
}

const skipped = (name) =>
  name.includes("/strings/") ||
  name.endsWith("/lib/say.ts") ||
  name.includes("__tests__") ||
  name.includes("/demo/");

// Tailwind, not English: every word carries a dash, a colon or a bracket.
const looksLikeClasses = (text) =>
  text.split(/\s+/).every((word) => /[-:[\]/]/.test(word) && !/[.,!?]$/.test(word));

function readable(text) {
  const trimmed = text.trim();
  if (trimmed.length < 3) return false;
  if (!/[A-Za-z]/.test(trimmed)) return false;
  if (trimmed.startsWith("/") || trimmed.includes("://")) return false;
  if (looksLikeClasses(trimmed)) return false;
  // Two words or more, beginning the way a sentence does, is prose. One
  // word is a value far more often than it is a label.
  return /\s/.test(trimmed) && /^[A-Z]/.test(trimmed);
}

for (const path of sources(src)) {
  const name = relative(root, path);
  if (skipped(name)) continue;

  const text = readFileSync(path, "utf8");
  const lines = text.split("\n");
  if (lines.slice(0, 20).some((line) => line.includes(MARKER))) continue;
  const file = ts.createSourceFile(path, text, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX);

  const lineOf = (node) =>
    file.getLineAndCharacterOfPosition(node.getStart()).line;

  const excused = (node) => {
    const line = lineOf(node);
    // The line itself, or either of the two above it: in JSX the marker
    // often has to sit outside the element that holds the text.
    return [0, 1, 2].some((back) => (lines[line - back] ?? "").includes(MARKER));
  };

  const complain = (node, said) => {
    if (excused(node)) return;
    problems.push(`${name}:${lineOf(node) + 1} says "${said.trim()}"; every string comes from src/strings, by key`);
  };

  const walk = (node) => {
    if (ts.isJsxText(node)) {
      const said = node.getText().replace(/\s+/g, " ").trim();
      if (said && /[A-Za-z]/.test(said) && said.length > 2) complain(node, said);
    } else if (ts.isStringLiteral(node) || ts.isNoSubstitutionTemplateLiteral(node)) {
      const parent = node.parent;
      const inSaid =
        ts.isJsxAttribute(parent) &&
        SAYS.has(parent.name.getText()) &&
        node.text.trim() !== "" &&
        !keys.has(node.text);
      if (inSaid || readable(node.text)) {
        if (inSaid || !isAValue(node)) complain(node, node.text);
      }
    }
    ts.forEachChild(node, walk);
  };
  walk(file);
}

// A string handed to a function, compared, or used as a key is read by the
// machine. An import is not a sentence either.
function isAValue(node) {
  const parent = node.parent;
  if (ts.isImportDeclaration(parent) || ts.isExportDeclaration(parent)) return true;
  if (ts.isBinaryExpression(parent)) return true;
  if (ts.isCallExpression(parent)) {
    // say("key") is the whole point; everything else handed to a function is
    // a value: a path, a header, a query.
    return parent.expression.getText() !== "say";
  }
  if (ts.isPropertyAccessExpression(parent) || ts.isElementAccessExpression(parent)) return true;
  if (ts.isCaseClause(parent)) return true;
  if (ts.isTypeReferenceNode(parent) || ts.isLiteralTypeNode(parent)) return true;
  return false;
}

// Every language says the same thing, and none of them may use a word the
// hand-off bans.
const BANNED = [
  "commit", "commits", "committed",
  "branch", "branches",
  "push", "pushed", "pushes",
  "pull", "pulled", "pulls",
  "merge", "merged", "merges",
  "repository", "repositories", "repo", "repos",
  "deploy", "deployed", "deploys", "deployment",
  "SHA", "HEAD",
  "conflict", "conflicts",
  // The same words in Dutch, so that a translation cannot let one back in.
  "commitgeschiedenis", "branche", "branches", "pushen", "pullen",
  "samenvoegen", "mergen", "repositoryn", "uitrollen", "uitgerold",
  "conflicten", "conflict",
];

// For implementers, and marked as such where it is shown.
const FOR_IMPLEMENTERS = /^(onboarding|panels|project-settings|sync|setup)\./;

for (const language of ["en", "nl"]) {
  const path = join(src, "strings", `${language}.ts`);
  const text = readFileSync(path, "utf8");
  const file = ts.createSourceFile(path, text, ts.ScriptTarget.Latest, true, ts.ScriptKind.TS);

  const walk = (node) => {
    if (ts.isPropertyAssignment(node) && ts.isStringLiteral(node.initializer)) {
      const key = node.name.getText().replace(/^"|"$/g, "");
      const said = node.initializer.text;
      if (FOR_IMPLEMENTERS.test(key)) return;
      for (const word of BANNED) {
        if (new RegExp(`\\b${word}\\b`, "i").test(said)) {
          problems.push(
            `strings/${language}.ts: ${key} says "${said}", which uses the word "${word}"; editors never see version control words`,
          );
        }
      }
    }
    ts.forEachChild(node, walk);
  };
  walk(file);
}

if (problems.length > 0) {
  console.error(`check-strings: ${problems.length} problem(s)`);
  for (const problem of problems) console.error(`  ${problem}`);
  process.exit(1);
}

console.log("check-strings: every string comes from the catalogue");
