Analyze the upstream changelog for the fork in `PATCHER.md` and produce a set of GitHub issues representing discrete, parallel units of backport work.

Arguments (optional): `$ARGUMENTS` — override the fork to dissect (e.g. `owner/repo`). If omitted, read from `PATCHER.md`.

---

## Step 1 — Load context

Read `PATCHER.md`. Extract:
- `upstream` repo
- `fork` repo
- `last_patched` version (e.g. `v1.2.3`)

If `last_patched` is not set, inspect the fork's git tags for the most recent `vX.Y.Z-patch` tag and derive the upstream base version from it.

## Step 2 — Enumerate upstream versions

List all releases published on the upstream repo **after** `last_patched`, in chronological order. For each:
- Retrieve release notes
- Retrieve the list of commits and the full diff introduced by that release

## Step 3 — Semantic analysis, version by version

For each version hop (`vA → vB`), reason through the diff at a feature level:

- Identify discrete features, bug fixes, refactors, and breaking changes
- Group related commits into named changes. If a feature is assembled from multiple commits or if one commit sets up something that a later commit builds on, treat these as a **story** — a parent feature with a sequence of sub-points
- Note any changes that touch areas described in `PATCHER.md` (custom features, architectural differences). Flag these as potential conflicts

Do not just summarize commit messages. Read the code. Understand what actually changed and why.

## Step 4 — Construct a unified changelog

Collapse all version analyses into a single feature-level changelog:

```
[feature name]
  upstream version: v1.4.0
  type: standalone | story
  description: ...
  upstream commits: ...
  fork conflict: yes/no — reason if yes
```

Order entries by: stories first (they often unblock standalone items), then standalone features, then fixes.

## Step 5 — Create GitHub issues

For each entry in the changelog, create one GitHub issue on the **fork** repository:

**Title**: `[backport] <feature name> (upstream <version>)`

**Labels**: `backport`, `patch-dissect`

**Body**:
```
## Summary
<what this upstream change does>

## Upstream reference
- version: <vX.Y.Z>
- commits: <links>
- PR/issue (if any): <link>

## Fork conflict
<none | description of what in PATCHER.md this touches and why>

## Context
<any additional semantic notes — why this change exists upstream, what problem it solves>
```

For **stories**, create one parent issue with a task list (`- [ ] #<sub-issue>`) and create the sub-issues individually.

For issues that **depend** on another issue being applied first, add `Depends on #N` to the body.

Finally, post a single summary comment on all created issues that contains the full unified changelog, so every issue has the full picture as context.
