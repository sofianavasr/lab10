---
name: defining-tdd-workflow
description: Define and enforce a practical TDD workflow for this codebase. Use when implementing features, fixing bugs, or refactoring where use cases must be explicit, tests must be written before implementation, and verification must confirm behavior.
---

# Defining TDD

## Purpose

Use this skill to run implementation work in strict test-first order:
1. Define behavior as explicit use cases.
2. Write failing tests first.
3. Implement only enough code to pass.
4. Refactor safely with tests green.
5. Verify behavior with targeted and full checks.

## Quick Rules

- Always follow **Red -> Green -> Refactor**.
- Never write production logic before a failing test exists for that behavior.
- Keep one behavioral change per cycle.
- Prefer table-driven tests with `t.Run()` for service/repository logic.
- Preserve architecture boundaries: `handler -> service -> repository`.
- For bug fixes, first reproduce the bug with a failing test.

## Workflow

Copy this checklist and keep it updated:

```markdown
TDD Progress
- [ ] 1) Define use cases and acceptance criteria
- [ ] 2) Write/adjust tests that fail for the new behavior
- [ ] 3) Implement minimal code to make tests pass
- [ ] 4) Refactor code/tests while keeping tests green
- [ ] 5) Run verification (targeted + full suite as needed)
- [ ] 6) Report what behavior is now guaranteed by tests
```

### 1) Define Use Cases

Before coding, capture:
- Primary use case ("when X, system does Y")
- Edge cases (invalid input, missing data, dependency errors)
- Expected API behavior (status code + JSON error message format when relevant)

If requirements are unclear, ask clarifying questions before writing code.

### 2) Red: Write Failing Tests First

Choose the lowest layer where behavior is best validated:
- **Service logic**: unit tests in `internal/service/...`
- **Repository behavior**: repository/query tests where applicable
- **Handler contracts**: handler tests for request/response behavior

Test guidance:
- Use table-driven tests where practical.
- Name tests by behavior, not implementation details.
- Fail for the right reason (assert observable behavior).

### 3) Green: Minimal Implementation

Write only the code needed to pass current failing tests:
- Do not add speculative features.
- Keep handlers thin; business logic stays in services.
- Wrap errors with context in service/repository code.
- Return safe, human-readable API errors from handlers.

### 4) Refactor Safely

After tests pass:
- Improve naming, extract helpers, simplify flow.
- Keep functions small and focused.
- Re-run changed tests after each refactor step.

### 5) Verify

Run targeted tests first, then broader checks proportional to risk:

```bash
# focused package
go test ./internal/service/... -v

# or a single test during iteration
go test ./internal/middleware/ -run TestName -v

# full safety checks before finishing substantial changes
make check
```

If SQL queries changed, regenerate and verify generated artifacts:

```bash
make sqlc
go test ./... -v
```

## Output Expectations

When reporting completion, include:
- Use cases implemented.
- Tests added/updated (and what they prove).
- Verification commands run and outcomes.
- Any remaining risks, assumptions, or follow-ups.

## Done Criteria

Work is done only when:
- New/changed behavior is covered by tests.
- Tests were written before implementation (or bug reproduction test came first).
- Relevant checks pass for touched areas.
- Code follows project conventions and architecture boundaries.
