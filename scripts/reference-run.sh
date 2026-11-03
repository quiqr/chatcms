#!/usr/bin/env bash
# Runs the old Quiqr's standalone server in its own dev shell. Its own file
# so that the process holding the port is the one whose pid was written down,
# rather than a shell that started another shell.
set -euo pipefail
app=$1
exec nix develop "$app" --no-warn-dirty --command bash -c '
    export LD_LIBRARY_PATH="$(nix eval --raw nixpkgs#stdenv.cc.cc.lib.outPath)/lib${LD_LIBRARY_PATH:+:$LD_LIBRARY_PATH}"
    cd "$1"
    exec node packages/adapters/standalone/dist/main.js
' -- "$app"
