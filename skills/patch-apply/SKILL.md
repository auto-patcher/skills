Apply all `ready` issues in dependency order. You are the **super-agent** (autopatcher). Sub-agents implement individual issues. Your job is to supervise, maintain context across the full patch cycle, and ensure every applied change is correct and stylistically sound.

Run this after `/patch-design`.

---

## Step 1 — Load context

Read `PATCHER.md`, including the Testing section. You will use it throughout this cycle.

List all open issues in the fork repo labeled `ready`. Skip issues labeled `conflict` — they require human resolution.

Sort by dependency order:
- Issues with no `Depends on` come first
- Respect story ordering (parent issue defines the sequence of sub-issues)

## Step 2 — Apply each issue

Work through the sorted list. For each issue:

### 2a — Brief the sub-agent

Spawn a sub-agent with the following context:

- The full issue body (what is being implemented)
- The design comment (how to implement it in fork style)
- The Testing section from `PATCHER.md` (commands, smoke tests, subagent testing setup)
- The instruction: **implement as described in the design comment, not as a copy of upstream code**. The sub-agent should write code as if it were a native member of this fork's codebase.

Sub-agent task:
1. Implement the changes described in the design comment
2. Run the existing test suite — fix any failures introduced
3. Write new tests for the implemented behavior
4. Execute the full testing plan from the design comment: unit tests, build verification, smoke tests, and any subagent testing scenarios. Use `PATCHER.md`'s Testing section for commands and expected outcomes.
5. Commit with message: `[patch] <issue title> (fixes #<issue-number>)`

### 2b — Review as super-agent

After the sub-agent reports back:

- Read the diff. Ask: does this look like it belongs in this fork, or does it look like it was pasted from somewhere else?
- Verify all tests passed — unit, build, smoke, and subagent tests
- Check that the fork's character and architecture from `PATCHER.md` are intact
- If the work needs changes: either request a revision from the sub-agent (for significant issues) or apply the fixes directly (for small style corrections)
- Once satisfied, close the issue with a comment summarizing what was done and linking the commit

### 2c — Track state

After each issue is applied and closed, note it as complete before moving to the next. If something unexpected comes up mid-cycle (a conflict, a test failure that reveals a deeper problem, a design assumption that was wrong), pause and surface it rather than pushing through.

## Step 3 — Finalize the patch cycle

Once all `ready` issues are applied and closed:

1. Run the **full test suite** on the patch branch — unit tests, build, all smoke tests from `PATCHER.md`'s Testing section, and any subagent testing it describes. Leave no angle untested.
2. If anything fails: do not proceed. Investigate and fix, then retry from step 1.
3. If all tests pass, open a pull request against `main` with a summary of all changes applied, grouped by type (`backport`, `feature`, `bug`). Include links to all closed issues.
4. Merge the pull request into `main`.
5. Run the **integration test suite** against `main` after merge.
6. If integration tests fail: do not release. Investigate on `main`, fix, and re-run integration tests before continuing.
7. If integration tests pass:
   - If this cycle included any `backport` issues: update `last_patched` in `PATCHER.md` on `main` to the highest upstream version covered by this cycle
   - Create a **GitHub release** tagged `v<upstream_version>-patch` (for backport cycles) or `v<fork_version>` (for pure feature/bug cycles) targeting `main`. The release body should summarize all changes by type.

## Principles

**Style over verbatim.** The fork should gain the capability in its own voice. If the sub-agent produces code that looks copied, send it back.

**Test from every angle.** Unit tests alone are not enough. Build it, run it, smoke test it, and use subagents where the Testing section calls for it. A release that hasn't been exercised end-to-end is not ready.

**Conflicts block the cycle.** If a sub-agent surfaces a conflict with the fork's character or architecture, stop and surface it for human review.

**Context is your job.** Sub-agents see one issue at a time. You see the whole cycle. Notice cross-issue concerns, revisit earlier assumptions when needed, keep the sum of changes coherent.

**Caution over speed.** There is no deadline. Ship correct, stylistically integrated work — not fast work.
