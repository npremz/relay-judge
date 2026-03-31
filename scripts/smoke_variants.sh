#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

BIN_PATH="$(mktemp "${TMPDIR:-/tmp}/relay-judge-smoke.XXXXXX")"
TMP_DIR="$(mktemp -d "${TMPDIR:-/tmp}/relay-judge-smoke-dir.XXXXXX")"

cleanup() {
  rm -f "$BIN_PATH"
  rm -rf "$TMP_DIR"
}
trap cleanup EXIT

go build -o "$BIN_PATH" ./cmd/relay-judge
cp ./examples/two_sum.py "$TMP_DIR/two-sum.py"

run_case() {
  local name="$1"
  local expected_code="$2"
  local expected_text="$3"
  shift 3

  local output
  local code

  set +e
  output="$("$@" 2>&1)"
  code=$?
  set -e

  if [[ "$code" -ne "$expected_code" ]]; then
    echo "[$name] expected exit $expected_code, got $code" >&2
    echo "$output" >&2
    exit 1
  fi

  if ! grep -Fq "$expected_text" <<<"$output"; then
    echo "[$name] missing expected text: $expected_text" >&2
    echo "$output" >&2
    exit 1
  fi

  echo "[ok] $name"
}

run_case "pass" 0 "PASSED" \
  "$BIN_PATH" run --subject two-sum --workspace ./examples

run_case "wrong" 1 "FAILED" \
  "$BIN_PATH" run --subject two-sum --workspace ./examples/variants/wrong

run_case "runtime" 2 "RUNTIME_ERROR" \
  "$BIN_PATH" run --subject two-sum --workspace ./examples/variants/runtime

run_case "syntax" 4 "LOAD_ERROR" \
  "$BIN_PATH" run --subject two-sum --workspace ./examples/variants/syntax

run_case "timeout" 3 "TIMEOUT" \
  "$BIN_PATH" run --subject two-sum --workspace ./examples/variants/timeout

run_case "stress-pass" 0 "Mode      : stress" \
  "$BIN_PATH" --stress "$TMP_DIR/two-sum.py"

run_case "stress-slow" 3 "TIMEOUT" \
  "$BIN_PATH" --stress ./examples/variants/slow/two_sum.py

run_case "run-trailing-stress" 0 "Mode      : stress" \
  "$BIN_PATH" run ./examples/two_sum.py --stress

echo "All smoke checks passed."
