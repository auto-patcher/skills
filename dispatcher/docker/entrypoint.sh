#!/usr/bin/env bash
# Runs a full patch cycle for a single fork repo and exits.
#
# Required environment:
#   TARGET_REPO       full repo name, e.g. "auto-patcher/my-repo"
#   GITHUB_TOKEN      PAT with repo + org read/write
#   ANTHROPIC_API_KEY

set -euo pipefail

REPO="${TARGET_REPO:-${1:-}}"
if [[ -z "$REPO" ]]; then
    echo "error: TARGET_REPO is required" >&2
    exit 1
fi

echo "[autopatcher] starting cycle for $REPO"

# Configure git and gh auth.
gh auth setup-git
git config --global user.email "autopatcher@auto-patcher.io"
git config --global user.name "Auto Patcher"

# Clone the fork into a clean workspace.
gh repo clone "$REPO" /workspace/repo -- --depth=50
cd /workspace/repo

# Run the full patch cycle non-interactively.
#
# The prompt instructs the agent to run the full cycle.
# CLAUDE.md in the repo provides agent identity; skills/ provides skill definitions.
#
# TODO: update the --print flag / invocation once the auto-patcher CC fork
# stabilises its non-interactive API. Key requirements:
#   - Non-interactive: no TTY expected
#   - Must be able to invoke skills (/patch-dissect, /patch-design, /patch-apply)
#   - Must exit non-zero on unrecoverable failure

claude --print "
You are the autopatcher. Run a full patch cycle for this repository.

1. Read PATCHER.md to understand this fork's identity, upstream repo, and last_patched version.
2. If the upstream has releases newer than last_patched: run /patch-dissect.
3. Run /patch-design on all eligible open issues.
4. Run /patch-apply on all ready issues.

Work methodically. Do not skip steps. Surface conflicts rather than forcing through them.
Exit when the cycle is complete or when you have surfaced an issue requiring human review.
"

echo "[autopatcher] cycle complete for $REPO"
