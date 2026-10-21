// Runs the reference serialisers over the case table and records their output.
// Invoked by scripts/generate-goldens.sh, never by the gate.
import { createRequire } from "node:module";
import { mkdirSync, readFileSync, writeFileSync } from "node:fs";
import { join } from "node:path";

const [workdir, casesPath, outDir] = process.argv.slice(2);

const require = createRequire(join(workdir, "noop.cjs"));
const yaml = require("js-yaml");
const tomlify = require("tomlify-j0.4");

const yamlVersion = require("js-yaml/package.json").version;
const tomlifyVersion = require("tomlify-j0.4/package.json").version;

// A date cannot be written in JSON, so the case table marks one with a prefix
// and it is revived here.
const DATE_PREFIX = "@@DATE@@";

function revive(value) {
  if (typeof value === "string" && value.startsWith(DATE_PREFIX)) {
    return new Date(value.slice(DATE_PREFIX.length));
  }
  if (Array.isArray(value)) {
    return value.map(revive);
  }
  if (value !== null && typeof value === "object") {
    return Object.fromEntries(Object.entries(value).map(([k, v]) => [k, revive(v)]));
  }
  return value;
}

// The three dumpContent implementations from
// ../quiqr-desktop/packages/backend/src/utils/format-providers/, reproduced
// here so the goldens cover the whole file and not only the front matter.
function dumpYamlContent(obj) {
  const copy = { ...obj };
  let content = "";
  if (copy.mainContent) {
    content = copy.mainContent;
    delete copy.mainContent;
  }
  if (Object.keys(copy).length === 0) {
    return `${content}`;
  }
  return `---\n${yaml.dump(copy)}---\n\n${content}`;
}

function dumpTomlContent(obj) {
  const copy = { ...obj };
  let content = "";
  if (copy.mainContent) {
    content = copy.mainContent;
    delete copy.mainContent;
  }
  return `+++\n${tomlify.toToml(copy, { space: 2 })}+++\n\n${content}`;
}

function dumpJsonContent(obj) {
  const copy = { ...obj };
  let content = "";
  if (copy.mainContent) {
    content = copy.mainContent;
    delete copy.mainContent;
  }
  return `${JSON.stringify(copy, null, 2)}\n\n${content}`;
}

const cases = JSON.parse(readFileSync(casesPath, "utf8"));
const index = [];

mkdirSync(outDir, { recursive: true });

for (const testCase of cases) {
  const value = revive(testCase.value);

  const yamlOut = yaml.dump(value);
  writeFileSync(join(outDir, `${testCase.name}.yaml.golden`), yamlOut);

  const documentOut = dumpYamlContent({ ...value, mainContent: "The body.\n" });
  writeFileSync(join(outDir, `${testCase.name}.document.golden`), documentOut);

  const entry = { name: testCase.name, yaml: true, document: true };

  // tomlify and JSON cannot express everything YAML can; a case they refuse is
  // recorded as refused rather than quietly skipped.
  try {
    writeFileSync(join(outDir, `${testCase.name}.toml.golden`), tomlify.toToml(value, { space: 2 }));
    entry.toml = true;
  } catch (error) {
    entry.toml = false;
    entry.tomlError = String(error.message ?? error);
  }

  try {
    writeFileSync(join(outDir, `${testCase.name}.json.golden`), `${JSON.stringify(value, null, 2)}\n`);
    entry.json = true;
  } catch (error) {
    entry.json = false;
    entry.jsonError = String(error.message ?? error);
  }

  index.push(entry);
}

writeFileSync(
  join(outDir, "index.json"),
  `${JSON.stringify({ jsYaml: yamlVersion, tomlify: tomlifyVersion, cases: index }, null, 2)}\n`,
);

console.log(`js-yaml ${yamlVersion}, tomlify-j0.4 ${tomlifyVersion}: ${cases.length} cases`);
