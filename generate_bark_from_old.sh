#!/usr/bin/env zsh
set -euo pipefail

go run bark.go reverse-gen "old/*.html"
