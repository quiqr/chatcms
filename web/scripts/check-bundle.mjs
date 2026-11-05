import { readdirSync, readFileSync, statSync } from "node:fs";
import { join, relative } from "node:path";

const root = new URL("..", import.meta.url).pathname;
const dist = process.argv[2] ?? join(root, "dist");

const forbidden = [
  "a long enough passphrase",
  "marieke@korenaar.example",
  "sam@korenaar.example",
  "owner@example.org",
  "admin@example.org",
];

function* files(dir) {
  for (const entry of readdirSync(dir)) {
    const path = join(dir, entry);
    if (statSync(path).isDirectory()) {
      yield* files(path);
    } else if (/\.(js|html|css|map)$/.test(entry)) {
      yield path;
    }
  }
}

const problems = [];
for (const path of files(dist)) {
  const text = readFileSync(path, "utf8");
  for (const word of forbidden) {
    if (text.includes(word)) problems.push(`${relative(root, path)} carries "${word}"`);
  }
}

if (problems.length > 0) {
  console.error("check-bundle: development sign-in reached the production build");
  for (const problem of problems) console.error(`  ${problem}`);
  process.exit(1);
}
console.log("check-bundle: no development sign-in in the production build");
