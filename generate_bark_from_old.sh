#!/usr/bin/env zsh
set -euo pipefail

for src in old/*.html; do
  out="$(basename "${src%.html}.bark")"
  python3 html_to_bark.py "$src" > "$out"
done
