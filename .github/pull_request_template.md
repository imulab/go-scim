<!--
  This template is auto-applied to every new PR. Fill in each section before
  requesting review. Anti-relapse self-attestation is a hard gate per
  CONTRIBUTING.md; the CI grep mirror catches violations regardless of how the
  checkboxes are answered.
-->

## What does this PR do?

<!-- One sentence. -->

## Why is this change needed?

<!-- Link to issue, PROJECT.md decision, or CONTEXT.md context. -->

## Anti-relapse self-attestation (per CONTRIBUTING.md)

- [ ] **Does this PR introduce IdP-specific tolerance** (Entra, Okta, Azure AD, etc.)?
      Answer: **No** (if Yes, this PR will be closed; see CONTRIBUTING.md Rule 3 and PITFALLS.md Pitfall 3).

- [ ] **Does this PR introduce a runtime-typed `Resource` interface or a generic `Property` tree**?
      Answer: **No** (if Yes, this PR will be closed; see CONTRIBUTING.md Rule 1 and PITFALLS.md Pitfall 1).

- [ ] **Does this PR add `os.ReadFile` of schema content at runtime** (or any non-`embed.FS` schema loading)?
      Answer: **No** (if Yes, this PR will be closed; see CONTRIBUTING.md Rule 2 and PITFALLS.md Pitfall 2).

## Tests

<!-- What tests were added, or what was verified manually. -->
