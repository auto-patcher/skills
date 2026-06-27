Review all open backport issues and annotate each with a design plan that expresses the upstream change in this fork's style.

Run this after `/patch-dissect` and before `/patch-apply`.

---

## Step 1 — Load context

Read `PATCHER.md`. Internalize:
- The fork's purpose and what makes it distinct
- Custom features that must be preserved
- Architectural patterns the fork uses
- Style notes

List all open issues in the fork repo labeled `backport` and `patch-dissect`. These are the units of work to design.

## Step 2 — Order by dependency

Before designing, determine the dependency order:
- Issues with no dependencies come first
- If issue B says "Depends on #A", design A before B (B's design may depend on how A was implemented)
- Stories: design the parent before the sub-issues

## Step 3 — Design each issue

For each issue, in dependency order:

1. **Read** the issue body: what does this upstream change do?
2. **Fetch** the relevant upstream code (commits and files referenced in the issue)
3. **Read** the corresponding area of the fork's codebase
4. **Think**: if a developer who wrote this fork were implementing this feature from scratch — without looking at upstream — how would they do it?

   Consider:
   - Does the upstream abstraction fit the fork's existing abstractions, or should it be expressed differently?
   - Are there naming patterns in the fork that should be followed?
   - Does the fork already have partial infrastructure for this that should be extended rather than worked around?
   - Can the upstream feature be implemented in a way that also benefits or extends the fork's custom features?
   - Is the upstream change's scope right, or should it be split or merged with existing code?

5. **Write a design comment** on the issue:

```
## Design [patch-design]

### Approach
<how to implement this in the fork's style — not a port, a reimplementation in fork voice>

### Files to touch
- `path/to/file.ext` — reason
- ...

### Risks
<anything that could break existing fork behavior; call out any tension with PATCHER.md custom features>

### Testing plan
<how to verify this works: existing tests to run, new tests to write, manual checks>

### Notes
<anything else relevant — e.g. if the upstream change is poor quality and we should do it differently anyway>
```

6. **Add label** `patch-design` to the issue.

## Step 4 — Close no-ops

If the fork already has equivalent functionality for an issue (e.g. it was independently implemented), close the issue with a comment explaining:
- What existing code provides the equivalent behavior
- Why no change is needed

Add label `no-op` before closing.

## Step 5 — Flag blockers

If designing an issue reveals that it cannot be safely implemented without first resolving a fork conflict noted in the issue body, add label `conflict` and post a comment describing the conflict clearly. Do not write a design comment — surface it for human review instead.
