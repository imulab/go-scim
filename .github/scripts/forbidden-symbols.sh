#!/usr/bin/env bash
set -euo pipefail

# Forbidden-symbol grep gate — mirrors the lefthook pre-commit hooks.
# See CONTRIBUTING.md for the three rules these patterns enforce.

EXCLUDES=(
  ':(exclude).planning/'
  ':(exclude).legacy/'
  ':(exclude)CONTRIBUTING.md'
  ':(exclude)PROJECT.md'
  ':(exclude)README.md'
  ':(exclude).github/scripts/forbidden-symbols.sh'
  ':(exclude).github/workflows/ci.yml'
  ':(exclude).github/pull_request_template.md'
  ':(exclude)lefthook.yml'
)

PATTERNS=(
  'os\.ReadFile.*[Ss]chema'      # Rule 2: runtime schema interpretation
  'interface\s*\{\s*Path\('       # Rule 1: Property-tree relapse signature
  'Property\s+interface'          # Rule 1: interface declaration form
  'Entra|Okta|Azure\s*AD|AzureAD' # Rule 3: vendor-name accommodation
)

failed=0
for pat in "${PATTERNS[@]}"; do
  if git grep -nE "$pat" -- "${EXCLUDES[@]}" 2>/dev/null; then
    echo "::error::Forbidden pattern matched: $pat (see CONTRIBUTING.md)"
    failed=1
  fi
done

if [ "$failed" -ne 0 ]; then
  echo "Forbidden-symbol grep gate failed. See CONTRIBUTING.md for the rules."
  exit 1
fi

echo "Forbidden-symbol grep gate: OK"
