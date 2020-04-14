#!/bin/bash

case "${1:-}" in
lint)
  golangci-lint run ./... -v
  ;;
unit)
  go test ./...
  ;;
codegen)
  operator-sdk generate crds
  operator-sdk generate k8s
  git status --porcelain 2>/dev/null| grep "^??" | wc -l
  ;;
*)
    echo " $0 [command]
Test Toolbox

Available Commands:
  lint
  unit
  codegen
" >&2
    ;;
esac

