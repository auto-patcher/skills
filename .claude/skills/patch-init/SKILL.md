Initialize a new autopatcher-managed fork of an upstream repository.

Arguments: `$ARGUMENTS` — the upstream to fork, as `owner/repo` or a full GitHub URL.

---

## Step 1 — Parse arguments

Extract `owner/repo` from `$ARGUMENTS`, stripping any `https://github.com/` prefix.

Derive:
- `upstream`: `owner/repo`
- `repo_name`: the repo part
- `fork`: `auto-patcher/<repo_name>`
- `local_path`: `~/code/auto-patcher/<repo_name>`

## Step 2 — Fork to auto-patcher org

Use the GitHub MCP `fork_repository` tool to fork `<upstream>` into the `auto-patcher` organization. If a fork already exists there, note it and continue.

## Step 3 — Clone locally

```bash
gh repo clone auto-patcher/<repo_name> ~/code/auto-patcher/<repo_name>
```

## Step 4 — Fetch upstream metadata

Using GitHub MCP tools, fetch the latest release tag from the upstream repo. This becomes the initial `last_patched` value — we consider the fork "caught up" to upstream as of the fork point.

## Step 5 — Scaffold autopatcher files

Push the following files to `auto-patcher/<repo_name>` via GitHub MCP tools.

### `CLAUDE.md`

Fetch from `auto-patcher/skills` and copy verbatim. This is the shared agent definition — do not modify it.

### `PATCHER.md`

Create a filled-out instance with the repositories and baseline populated, and the identity sections left as prompts for the user:

```markdown
# Patcher

## Repositories

\`\`\`
upstream: <upstream>
fork:     auto-patcher/<repo_name>
\`\`\`

## Upstream baseline

\`\`\`
last_patched: <latest_upstream_release_tag>
\`\`\`

The upstream version tag last incorporated into this fork. The next `/patch-dissect` run will analyze everything released after this.

## Purpose

<!-- Why does this fork exist? What does it do that upstream doesn't, or differently? -->

## Character

<!-- How does this fork think and work? What trade-offs does it make differently from upstream?
     What would a developer native to this codebase care about? -->

## Architecture

<!-- Where does this fork diverge structurally? Different modules, replaced dependencies,
     new abstractions, removed subsystems? Be concrete — name the packages or files that differ. -->

## Style

<!-- Naming conventions, idioms, and patterns specific to this fork. -->
```

### `.claude/skills/`

Copy all skills from `auto-patcher/skills` into the fork:
- `.claude/skills/patch-dissect/SKILL.md`
- `.claude/skills/patch-design/SKILL.md`
- `.claude/skills/patch-apply/SKILL.md`

## Step 6 — Nix integration

### If the fork has no `flake.nix`

Create one that provides a devShell with the skills symlinked in:

```nix
{
  description = "<repo_name> — auto-patcher fork of <upstream>";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
    skills = {
      url = "github:auto-patcher/skills";
      inputs.nixpkgs.follows = "nixpkgs";
      inputs.flake-utils.follows = "flake-utils";
    };
  };

  outputs = { self, nixpkgs, flake-utils, skills }:
    flake-utils.lib.eachDefaultSystem (system:
      let pkgs = nixpkgs.legacyPackages.${system}; in
      {
        devShells.default = pkgs.mkShell {
          shellHook = ''
            skill_src="${skills.packages.${system}.default}/share/claude/commands"
            mkdir -p .claude/commands
            for f in "$skill_src"/*; do
              ln -sf "$f" ".claude/commands/$(basename $f)"
            done
          '';
        };

        formatter = pkgs.nixfmt;
      }
    );
}
```

### If the fork already has a `flake.nix`

Add `auto-patcher/skills` as an input (with `nixpkgs` and `flake-utils` following the fork's existing inputs), then add the skills symlink shellHook to the existing `devShells.default`.

Push the updated `flake.nix` via GitHub MCP tools.

## Step 7 — Report

Summarize what was done:
- Fork URL: `https://github.com/auto-patcher/<repo_name>`
- Local clone: `~/code/auto-patcher/<repo_name>`
- Initial `last_patched`: `<tag>`
- Files added: `CLAUDE.md`, `PATCHER.md`, `.claude/skills/`, `flake.nix` (created or updated)

Remind the user to fill in the **Purpose**, **Character**, **Architecture**, and **Style** sections of `PATCHER.md` before running `/patch-dissect`. The agent cannot make good judgements about what to preserve or how to express changes without this context.
