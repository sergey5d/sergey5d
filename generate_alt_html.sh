#!/usr/bin/env zsh
set -euo pipefail

for src in *.bark; do
  case "$src" in
    sample.bark)
      continue
      ;;
  esac
done
go run bark.go gen "*.bark"
