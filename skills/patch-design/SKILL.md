Design the implementation for open issues on this fork. For each eligible issue, produce a design comment describing how to implement the request in this fork's style, then mark it `ready`.

An issue is eligible if:
- It is open
- It was opened by a member of the `auto-patcher` org
- It does not already have the `ready` label
- It is not labeled `conflict`

This applies to any issue type — `backport`, `feature`, `bug` — the design process is the same: understand the request, then figure out how this fork would express it.

Run this after `/patch-dissect` for backport cycles, or at any time for original feature and bug issues.

---

## Step 1 — Load context

Read `PATCHER.md`. Internalize:
- The fork's purpose, character, architecture, and style
- The Testing section — you will use this when writing testing plans

List all open, eligible issues (see criteria above). Check issue authors against the `auto-patcher` org member list.

## Step 2 — Order by dependency

- Issues with no dependencies come first
- If issue B says "Depends on #A", design A before B
- Story parent issues before their sub-issues

## Step 3 — Design each issue

For each eligible issue, in dependency order:

1. **Read** the issue body: what is being asked for?
2. If the issue is a `backport`: fetch the relevant upstream code (commits and files referenced in the issue). If it is a `feature` or `bug`: read the description and any linked context carefully.
3. **Read** the corresponding area of the fork's codebase.
4. **Think**: how would a developer who wrote this fork implement this from scratch?

   Consider:
   - Does the proposed abstraction fit the fork's existing patterns, or should it be expressed differently?
   - Are there naming conventions in the fork that apply?
   - Does the fork already have partial infrastructure that should be extended rather than worked around?
   - Can this be implemented in a way that also strengthens the fork's custom features?
   - Is the scope right, or should it be split or merged with existing code?

5. **Write a design comment** on the issue:

```
## Design

### Approach
<how to implement this in the fork's style>

### Files to touch
- `path/to/file.ext` — reason
- ...

### Risks
<anything that could break existing fork behavior; any tension with PATCHER.md character or architecture>

### Testing plan
<cover all angles from PATCHER.md's Testing section that are relevant to this change:
which unit tests to run or write, integration test scenarios, build verification,
specific smoke test steps, and any subagent testing that makes sense for this feature>

### Notes
<anything else — e.g. if a backport's upstream implementation is poor and we should do better>
```

6. **Add label** `ready` to the issue.

## Step 4 — Close no-ops

If the fork already has equivalent functionality (e.g. a backport of something already independently implemented), close the issue with a comment explaining what existing code covers it. Add label `no-op` before closing.

## Step 5 — Flag conflicts

If an issue cannot be safely implemented without resolving a conflict with the fork's character or architecture, add label `conflict` and post a comment describing the conflict clearly. Do not write a design comment. Surface it for human review.
