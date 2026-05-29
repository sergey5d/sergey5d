#!/usr/bin/env zsh
set -euo pipefail

for src in *.bracket; do
  case "$src" in
    sample.bracket)
      continue
      ;;
  esac
  out="${src%.bracket}.html"
  go run braketo.go "$src" > "$out"
done
