Apply all designed backport patches in dependency order. You are the **super-agent** (autopatcher). Sub-agents implement individual issues. Your job is to supervise, maintain context across the full patch cycle, and ensure every applied change is correct and stylistically sound.

Run this after `/patch-design`.

---

## Step 1 — Load context

Read `PATCHER.md`.

List all open issues in the fork repo labeled both `backport` and `patch-design`. These are ready to implement. Skip issues labeled `conflict` — they require human resolution.

Sort by dependency order:
- Issues with no `Depends on` come first
- Respect story ordering (parent issue defines the sequence of sub-issues)

## Step 2 — Apply each issue

Work through the sorted list. For each issue:

### 2a — Brief the sub-agent

Spawn a sub-agent with the following context:

- The full issue body (what the upstream feature does)
- The design comment from `/patch-design` (how to implement it in fork style)
- The relevant sections of `PATCHER.md` (custom features, architectural differences, style notes)
- The instruction: **implement the feature as described in the design comment, not as a copy of the upstream code**. The sub-agent should write code as if it were a native member of this fork's codebase.

Sub-agent task:
1. Implement the changes described in the design comment
2. Run the existing test suite — fix any failures introduced
3. Write new tests for the implemented behavior
4. Commit with message: `[patch] <issue title> (fixes #<issue-number>)`

### 2b — Review as super-agent

After the sub-agent reports back:

- Read the diff. Ask: does this look like it belongs in this fork, or does it look like upstream code was pasted in?
- Verify tests pass (all of them, not just the new ones)
- Check that no custom features from `PATCHER.md` were accidentally modified or broken
- If the work needs changes: either request a revision from the sub-agent (for significant issues) or apply the fixes directly (for small style corrections)
- Once satisfied, close the issue with a comment summarizing what was done and linking the commit

### 2c — Track state

After each issue is applied and closed, note it as complete before moving to the next. If something unexpected comes up mid-cycle (a conflict, a test failure that reveals a deeper problem, a design assumption that was wrong), pause and surface it rather than pushing through.

## Step 3 — Finalize the patch cycle

Once all issues are applied and closed:

1. Run the **full test suite** one final time on the current branch
2. If all tests pass:
   - Create a tag: `v<upstream_version>-patch` (use the highest upstream version covered by this cycle)
   - Update `PATCHER.md`: set `last_patched` to the new upstream version and append a row to the patch history table
   - Open a pull request with a summary of all changes applied, grouped by feature
3. If tests fail: do not tag. Investigate and fix, then retry step 3.

## Principles

**Style over verbatim.** The fork should gain the capability, not the upstream's implementation. If the sub-agent produces code that looks like it was copied, send it back.

**Test before closing.** No issue closes without passing tests.

**Conflicts block the cycle.** If a sub-agent surfaces a conflict with a fork custom feature, do not work around it silently. Stop, document the conflict, and surface it for human review.

**Context is your job.** Sub-agents see one issue at a time. You see the whole cycle. Notice when issues affect each other, when assumptions made early in the cycle need revisiting, when the sum of changes is diverging from the fork's identity.

**Caution over speed.** There is no deadline. A patch cycle can span multiple sessions. Ship correct, stylistically integrated work — not fast work.
