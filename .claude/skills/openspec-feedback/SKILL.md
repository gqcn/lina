---
name: openspec-feedback
description: >-
  Organize, fix, verify, and cover with tests any bugs, improvements, or gaps reported during verification of an OpenSpec change.
  Use this skill when the user reports bugs, issues, defects, problems, missing test cases, missing coverage,
  improvement points, or UX feedback — whether discovered during manual verification, testing, or daily use.
  Trigger phrases include (but are not limited to): "bug", "issue", "fix", "feedback", "问题反馈", "问题",
  "反馈", "缺陷", "修复", "改进", "优化", "改善", "功能改进", "改进点", or numbered/bulleted
  lists of problems or improvements (中文或英文编号列表均可触发).
  This skill ensures every reported issue is tracked as a task artifact
  before any fix begins, and that each fix is covered by E2E test cases following the openspec-e2e conventions.
  IMPORTANT: This skill MUST be activated whenever the user provides any form of bug report, issue list,
  improvement suggestion, or feedback — regardless of language. If the message contains multiple numbered
  items describing problems or desired changes, this skill takes priority over direct implementation.
license: MIT
compatibility: Requires openspec CLI.
metadata:
  author: gqcn
  version: "1.0"
---

# Feedback: Structured Fix, Verification & Test Coverage Loop

When users perform manual verification after AI-driven implementation, they often discover bugs or improvement points. This skill captures those issues, organizes them into a traceable task list stored in the change's `tasks.md`, systematically fixes and verifies each one, and ensures every fix is covered by E2E test cases — creating a closed-loop management process with regression protection.

The core principles:
1. **Spec is the source of truth** — If an issue involves a requirement-level change (missing behavior, behavioral change, requirement gap), update the delta specs first, then record the task.
2. **Write it down first, then fix it** — Every issue gets recorded as a task artifact before any code change happens. This creates accountability, traceability, and a clean audit trail.
3. **Every fix deserves a test** — After fixing an issue, evaluate whether a new E2E test case (or sub-assertion) is needed to prevent regression. If the fix changes user-observable behavior, a test case is required.

---

## When This Skill Activates

- User reports one or more bugs, defects, improvement points, or gaps (missing features / missing test cases / incomplete coverage)
- User describes untested scenarios, missing test cases, or test coverage gaps
- The project uses OpenSpec with an active change
- The issues relate to an existing implementation (post-development feedback)

---

## Workflow

### 1. Identify the Active Change

Determine which change the issues relate to:

- If the user specifies a change name, use it
- If conversation context makes it obvious, use that
- If only one active change exists, auto-select it
- If ambiguous, run `openspec list --json` and ask the user to select

Announce: "Applying feedback fixes to change: **<name>**"

### 2. Read Current Context

Read existing artifacts to understand the implementation state:

- Read `tasks.md` to understand the current task structure, naming conventions, and numbering
- Read `design.md` and `proposal.md` if they exist, for architectural context
- Read existing delta specs under `specs/` in the change directory
- Scan relevant source files mentioned by the user
- **Scan existing E2E test files** (`find hack/tests/e2e -name 'TC*.ts' | sort`) to understand current test coverage and the highest TC ID

This context is essential — fixes must be consistent with the existing architecture, coding patterns, and spec definitions.

### 3. Analyze and Organize Issues

Parse the user's reported issues carefully. For each issue:

1. **Classify** — Determine both the issue type and its spec impact level:
   - **Issue type**: bug (incorrect behavior), missing feature, UX improvement, test gap (missing test case / incomplete test coverage), or missing implementation
   - **Spec impact level**:
     - **Implementation-level** (spec is correct, code is wrong) — e.g., a function doesn't follow the existing spec, a typo in logic, a missing error check. *No spec update needed.*
     - **Spec-level** (requirement is missing, incomplete, or needs to change) — e.g., a scenario not covered by any spec, a behavioral change, a new user-facing requirement. *Spec update needed before task recording.*
     - **Internal optimization** (no user-observable behavior change) — e.g., performance improvement, code cleanup, internal refactor. *No spec update needed.*
2. **Identify root cause** — What's the likely technical root cause? Which files are affected?
3. **Assess impact** — What does this break? What's the blast radius?
4. **Define verification** — How will we confirm the fix works?
5. **Evaluate test coverage** — Does an existing E2E test already cover this scenario? If not, plan a new test case or sub-assertion.

Group related issues together. If one root cause explains multiple symptoms, merge them into a single task with multiple verification points.

### 4. Update Delta Specs (for Spec-Level Issues)

For issues classified as **spec-level** in Step 3, update the corresponding delta spec files **before** writing tasks. This ensures the specs remain the single source of truth and tasks are derived from specs, not the other way around.

**Workflow:**

1. Identify which capability's spec file is affected (under `specs/<capability>/spec.md` in the change directory)
2. Determine the delta operation type:
   - **ADDED Requirements** — A completely new requirement or scenario not previously specified
   - **MODIFIED Requirements** — An existing requirement whose behavior needs to change (copy the full original requirement block, then edit)
   - **REMOVED Requirements** — A requirement to be deprecated (include Reason and Migration)
3. Update the spec file following the existing format conventions:
   - Each requirement: `### Requirement: <name>` followed by description using SHALL/MUST
   - Each scenario: `#### Scenario: <name>` with WHEN/THEN format
   - Every requirement MUST have at least one scenario
4. If the issue spans a new capability not covered by any existing spec file, create a new `specs/<new-capability>/spec.md`

**Skip this step** for issues classified as implementation-level or internal optimization — they don't change the specs.

Announce which spec files were updated (if any) before proceeding to task recording.

### 5. Write the Task List to tasks.md

Append a new **Feedback section** to the existing `tasks.md` file. Follow the existing file's conventions for formatting and numbering.

**Section format:**

If `tasks.md` does not yet have a Feedback section, append one:

```markdown
## Feedback

- [ ] **FB-1**：<one-sentence description of the problem or improvement>
- [ ] **FB-2**：<one-sentence description of the problem or improvement>
```

Each task is a **single line** — a concise, direct description of the problem. Do NOT add sub-fields like 现象、根因、影响、测试 etc. Keep `tasks.md` lean. All analysis and root cause investigation happens during the fix phase, not in the task record.

If the Feedback section already exists (from a previous round), simply append new tasks with the next sequential number.

**Numbering rules:**
- Task IDs use simple sequential numbering: `FB-1`, `FB-2`, `FB-3`, ...
- Check the existing Feedback section for the last used number and continue from there
- All feedback tasks — regardless of when they were reported — live in the same single section

**Test coverage evaluation (internal analysis, not written to tasks.md):**
- For each issue, internally assess whether an existing E2E test covers the scenario or a new one is needed.
- If the fix changes **user-observable behavior**, a test case or sub-assertion is **required**.
- If the fix is **internal-only** (no UI change), a test case is **optional**.
- New test cases follow the **openspec-e2e** skill conventions.
- Prefer adding **sub-assertions to existing TC files** when the scenario naturally belongs to an existing test case's scope.

**Important:** Show the draft task list to the user and confirm before writing to `tasks.md`. The user may want to adjust priorities, merge issues, or add details. Once confirmed, append the section to the file.

### 6. Execute Fixes (Loop)

Work through the task list sequentially. For each task:

**a. Announce**
```
## Fixing FB-X: <issue title>
```

**b. Investigate**
- Read the relevant source files
- Understand the current behavior
- Confirm the root cause matches the analysis

**c. Implement the fix**
- Make minimal, focused changes
- Keep the fix scoped to the specific issue
- Follow existing code patterns and conventions
- If the fix reveals a deeper issue, pause and discuss with the user

**d. Write or update E2E test cases**
- Follow **openspec-e2e** conventions strictly

**e. Verify — MANDATORY before marking complete**
- Run the newly added or updated E2E test cases and confirm they **pass**
- If the project has a broader e2e suite, run it to check for regressions
- Check for side effects in related functionality
- **A task MUST NOT be marked complete until its corresponding E2E test(s) have been executed and passed.** If no E2E test is applicable (internal optimization), the fix must at least be verified by running the existing test suite without regressions.

**f. Update tasks.md**
- Mark the task as complete: `- [ ]` → `- [x]` — **only after step (e) passes**
- Never mark a task complete based solely on code changes without test verification

**g. Continue to next task**

### 7. Run Comprehensive Verification

After all individual fixes are complete:

1. Run the full test suite if available
2. Report results — which tests pass, which fail
3. If new failures appear, analyze whether they are regressions from the fixes
4. If regressions exist, add them as new tasks and loop back to Step 6

### 8. Report Completion

Display a summary:

```
## Feedback Complete

**Change:** <change-name>
**Issues reported:** X
**Issues fixed:** Y/X
**Tests added:** Z new test cases / sub-assertions
**Verification:** <all passed / N issues remaining>

### Fixed This Session
- [x] FB-1: <title> ✓ (test: TC0010a)
- [x] FB-2: <title> ✓ (test: 已有覆盖)
- [x] FB-3: <title> ✓ (test: TC0010b)

### Remaining (if any)
- [ ] FB-4: <title> — blocked by <reason>

All fixes verified and test-covered. The tasks.md has been updated with full fix records.
```

If all tasks are complete and verified, suggest archiving the change.

---

## Handling Edge Cases

**User reports a single issue:** Still follow the full workflow — even one issue benefits from being recorded before fixing. The task list will just have one item.

**User reports missing test cases:** This is a test gap, not a bug. Classify as test gap, record the expected test scenarios in the task, implement the test cases (add to e2e test scripts or unit tests as appropriate), then verify by running the tests. The fix is the new test code itself.

**Fix reveals additional problems:** Add them as new tasks in the same Feedback section. Announce: "While fixing FB-X, I discovered an additional issue. Adding FB-Y to the task list." If the new issue is spec-level, update the spec first before adding the task.

**Issue is actually a design change:** If a reported "bug" is actually a requirement change or design change rather than an implementation bug, classify it as spec-level. Update the delta specs first (Step 4), then record the task (Step 5). If the change is large enough to affect `design.md` (e.g., new API endpoints, new DB schema, architectural changes), discuss with the user whether to also update `design.md` before proceeding.

**No active openspec change:** If the project uses openspec but there's no active change (e.g., all archived), create a new feedback-specific change:
```bash
openspec new change "feedback-<brief-description>"
```
Then generate the tasks.md in that new change directory.

**Multiple rounds of feedback:** All feedback tasks from every round are appended to the same single Feedback section. Sequential numbering (`FB-1`, `FB-2`, ...) naturally preserves the chronological order of when issues were discovered and fixed.

**Test not feasible:** Some fixes (e.g., timing-sensitive race conditions, infrastructure-only changes) may not be practically testable via E2E. In such cases, verify by running the existing full test suite without regressions, and note the reason in the completion summary. The task can still be marked complete if the full suite passes.

---

## Guardrails

- **Specs before tasks for requirement-level changes** — If an issue changes user-observable behavior or adds missing requirements, update the delta spec first, then record the task. Specs are the source of truth; tasks are derived from specs.
- **Always write tasks before fixing** — Never start coding a fix without first recording it in tasks.md.
- **Confirm the task list with the user** — The user knows what they observed; validate your analysis matches their experience.
- **Minimal fixes** — Don't refactor or improve code beyond what's needed to fix the reported issue.
- **Every user-visible fix needs a test** — If the fix changes behavior the user can observe, write a test. No exceptions unless technically infeasible.
- **Follow openspec-e2e conventions** — All new test cases MUST follow the TC ID allocation, naming, POM, and fixture conventions defined in the openspec-e2e skill.
- **Verify each fix individually** — Don't batch all fixes and hope for the best.
- **No green check without green tests** — A task can only be marked `[x]` after its E2E test(s) have been executed and passed. Code changes alone are never sufficient to mark a task complete.
- **Update tasks.md in real time** — Mark tasks complete immediately after verification, not at the end.
- **Preserve existing task format** — Match the conventions already used in the file.
- **Match the language of the target file** — When appending to or updating an artifact (specs, tasks.md, etc.), use the same natural language as the existing content in that file. If the file is written in Chinese, write in Chinese; if in English, write in English. Do not mix languages within a single file.
- **Don't over-spec implementation bugs** — If the existing spec already describes the correct behavior and the code simply doesn't follow it, fix the code. Adding redundant spec entries creates noise.
- **Don't lose context** — If the user's description is detailed, preserve those details in the task record.
