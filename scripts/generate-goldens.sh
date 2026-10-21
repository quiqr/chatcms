#!/usr/bin/env bash
# generate-goldens.sh
#
# Produces the expected bytes for the serialisation tests by running the same
# packages the Quiqr desktop app runs, rather than by writing down what we
# believe they do. The difference matters: the whole point of these goldens is
# to find out where that belief is wrong.
#
# Run this by hand when a case is added to scripts/goldens/cases.json. It needs
# the network the first time, to install the two reference packages, so it is
# never part of `nix flake check`. Its output is committed.
set -euo pipefail

ROOT="$(git rev-parse --show-toplevel)"
cd "$ROOT"

WORK="${CHATCMS_GOLDEN_WORKDIR:-$ROOT/.golden-tools}"
OUT="$ROOT/test/fixtures/content/golden"

# The versions the Quiqr desktop app depends on. Changing either changes the
# bytes a customer's repository will hold, so they are pinned here and
# recorded in the output.
JS_YAML="js-yaml@4.1.0"
TOMLIFY="tomlify-j0.4@3.0.0"

mkdir -p "$WORK" "$OUT"

if [[ ! -d "$WORK/node_modules" ]]; then
  echo "==> installing the reference packages into $WORK"
  (cd "$WORK" && npm install --silent --no-audit --no-fund --no-package-lock "$JS_YAML" "$TOMLIFY")
fi

echo "==> generating goldens into $OUT"
node "$ROOT/scripts/goldens/generate.mjs" "$WORK" "$ROOT/scripts/goldens/cases.json" "$OUT"

echo "==> done; review the diff before committing it"
