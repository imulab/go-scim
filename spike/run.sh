#!/usr/bin/env bash
# Spike runner: builds and runs both emission-engine candidates, asserts their
# outputs are equivalent (modulo the engine-attribution header line), and
# proves deterministic regen by running each engine twice and asserting
# byte-identical output across runs (REQ-GEN-06 dry run).
#
# Run from repo root: ./spike/run.sh
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TEMPLATE_DIR="${REPO_ROOT}/spike/template"
JENNIFER_DIR="${REPO_ROOT}/spike/jennifer"
TEMPLATE_OUT="${TEMPLATE_DIR}/output/user_gen.go"
JENNIFER_OUT="${JENNIFER_DIR}/output/user_gen.go"

run_engine() {
  local dir="$1"
  local label="$2"
  local out="${dir}/output/user_gen.go"

  # First run: capture as the baseline for the determinism check.
  ( cd "$dir" && GOWORK=off go run . )
  cp "$out" "${out}.first"

  # Second run: must produce byte-identical output (REQ-GEN-06 dry run).
  ( cd "$dir" && GOWORK=off go run . )
  if ! diff -q "${out}.first" "$out" >/dev/null; then
    echo "Spike FAIL: ${label} is non-deterministic across consecutive runs:" >&2
    diff -u "${out}.first" "$out" >&2 || true
    rm -f "${out}.first"
    exit 1
  fi
  rm -f "${out}.first"
}

# spike/jennifer needs its module cache populated before `go run` succeeds in
# environments where go.sum entries haven't been fetched yet.
( cd "$JENNIFER_DIR" && GOWORK=off go mod tidy )

run_engine "$TEMPLATE_DIR" "spike/template"
run_engine "$JENNIFER_DIR" "spike/jennifer"

# Cross-engine equivalence: both engines must produce equivalent Go modulo
# the engine-attribution header line. The byte-comparison gate is the
# apples-to-apples contract from CONTEXT.md.
if ! diff -u -I '^// Code generated' "$TEMPLATE_OUT" "$JENNIFER_OUT"; then
  echo "Spike FAIL: outputs differ (see diff above)." >&2
  exit 1
fi

echo "Spike OK: both engines produce equivalent output, both deterministic across runs."
