#!/usr/bin/env zsh
set -euo pipefail

for src in alt/*.bracket; do
  out="${src%.bracket}.html"
  go run braketo.go "$src" > "$out"
done
