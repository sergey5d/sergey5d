#!/usr/bin/env zsh
set -euo pipefail

for src in *.bark; do
  case "$src" in
    sample.bark)
      continue
      ;;
  esac
  out="${src%.bark}.html"
  go run braketo.go "$src" > "$out"
done
