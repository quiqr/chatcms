#!/bin/sh
set -e

if grep -r "TODO" content/ >/dev/null 2>&1; then
    echo "A page still says TODO. Finish it before publishing."
    exit 1
fi
echo "The content looks finished."
