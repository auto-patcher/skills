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

Before writing `PATCHER.md`, **read the codebase** to generate meaningful content rather than leaving placeholders. Use the GitHub MCP tools to explore the fork:

1. Fetch the repository tree to understand the top-level structure
2. Read key files: `README.md`, `CHANGELOG.md`, any docs directory, build files (`Makefile`, `flake.nix`, `package.json`, `Cargo.toml`, `go.mod`, etc.), and CI configuration (`.github/workflows/`)
3. Sample source files from the main language directories to understand naming conventions, idioms, and patterns
4. Look for test files and test runner configuration to understand the testing setup
5. Read the upstream `README.md` and compare with the fork's — differences reveal purpose and divergence

From this analysis, produce a filled-out `PATCHER.md`. Write real sentences based on what you found. Do not leave placeholder comments — if you genuinely cannot determine something (e.g. the fork has no README and no docs), write a brief note explaining what is unknown rather than a generic prompt.

```markdown
# Patcher

## Repositories

```
upstream: <upstream>
fork:     auto-patcher/<repo_name>
```

## Upstream baseline

```
last_patched: <latest_upstream_release_tag>
```

The upstream version tag last incorporated into this fork. The next `/patch-dissect` run will analyze everything released after this.

## Purpose

<derived from README diff, fork description, and any docs — what this fork does differently or why it exists>

## Character

<derived from reading the code — what trade-offs does this fork make, what does a developer native to it care about>

## Architecture

<derived from the repo tree and source files — concrete structural divergences from upstream: renamed packages, replaced dependencies, removed or added subsystems, key files that differ>

## Style

<derived from sampling source files — naming conventions, idioms, formatting patterns specific to this fork>

## Testing

### Unit tests

<command derived from Makefile / CI / test runner config>

### Integration tests

<command and required setup derived from CI or docs; note "not found" if absent>

### Build

<command derived from build files>

### Smoke tests

<concrete manual checks derived from README or docs; note "not determined" if absent>

### Subagent testing

<any scenarios that benefit from a subagent, e.g. acting as a client to a server — derived from the project's nature>
```

### `.claude/skills/`

Copy all skills from `auto-patcher/skills` into the fork:
- `.claude/skills/patch-dissect/SKILL.md`
- `.claude/skills/patch-design/SKILL.md`
- `.claude/skills/patch-apply/SKILL.md`

Fetch each from `auto-patcher/skills` at path `skills/<name>/SKILL.md`.

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
            skill_src="${skills.packages.${system}.default}/share/claude/skills"
            mkdir -p .claude/skills
            for f in "$skill_src"/*; do
              ln -sf "$f" ".claude/skills/$(basename $f)"
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

Review the generated `PATCHER.md` with the user. Highlight any sections where you had low confidence — missing docs, sparse README, ambiguous test setup — so they know what to verify or correct before running `/patch-dissect`. The Testing section in particular must be accurate, since all three operational skills depend on it.
