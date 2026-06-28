{
  description = "auto-patcher Claude skills and agents";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";

    # The claude-compatible coding agent the patcher drives per repo. Building
    # its root derivation yields the CLI. Fetching this input needs github.com
    # access — configure nix `access-tokens` (the workflow does this with the
    # built-in GITHUB_TOKEN).
    free-code.url = "github:gastrodon/free-code";
  };

  outputs =
    {
      self,
      nixpkgs,
      flake-utils,
      free-code,
    }:
    let
      # Package only the SKILL.md prompt files for installation.
      # Excludes Go source files (embed.go) and anything else alongside the prompts.
      # Downstream flakes can symlink $out/share/claude/skills/* into
      # ~/.claude/skills/ to make skills available as slash commands.
      mkSkillsPackage =
        pkgs:
        pkgs.runCommand "auto-patcher-skills" { } ''
          mkdir -p $out/share/claude/skills
          for skill in ${./skills}/*/SKILL.md; do
            name=$(basename $(dirname $skill))
            mkdir -p $out/share/claude/skills/$name
            cp $skill $out/share/claude/skills/$name/SKILL.md
          done
        '';

      # Package the autopatcher agent files (CLAUDE.md + PATCHER.md template).
      mkAgentPackage =
        pkgs:
        pkgs.runCommand "auto-patcher-agent" { } ''
          mkdir -p $out/share/auto-patcher
          cp ${./CLAUDE.md} $out/share/auto-patcher/CLAUDE.md
          cp ${./PATCHER.md} $out/share/auto-patcher/PATCHER.md
        '';

      # Build the auto-patcher binary.
      # It embeds the skill prompts at compile time via //go:embed and is run
      # for a single scan-and-drain pass by the auto-patch GitHub workflow.
      # Prerequisite: run `go mod tidy` at the repo root to generate go.sum,
      # then update vendorHash with the hash from the first failed `nix build`.
      mkAutoPatcherPackage =
        pkgs:
        pkgs.buildGoModule {
          pname = "auto-patcher";
          version = "0.1.0";
          src = ./.;
          subPackages = [ "." ];
          # Placeholder: run `nix build .#auto-patcher` after `go mod tidy` to
          # get the real hash from the error output, then replace this value.
          vendorHash = "sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=";
        };
    in
    {
      lib = {
        inherit mkSkillsPackage mkAgentPackage mkAutoPatcherPackage;
      };
    }
    // flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        # free-code's root derivation: the claude-compatible agent CLI.
        freeCodeCli = free-code.packages.${system}.default;
      in
      {
        packages = {
          default = mkSkillsPackage pkgs;
          agent = mkAgentPackage pkgs;
          auto-patcher = mkAutoPatcherPackage pkgs;
        };

        # The patch-run environment. Kept deliberately separate from the
        # auto-patcher binary derivation above: this is a `nix develop` shell,
        # not a package. It carries everything a run needs — the Go toolchain to
        # build the CLI, GNU parallel as the worker pool, git, and the free-code
        # agent the runner invokes. The auto-patch workflow enters this shell
        # inside a Lix container, so "setup" lives here in the flake rather than
        # as a pile of imperative YAML steps.
        devShells.default = pkgs.mkShell {
          packages = [
            pkgs.go
            pkgs.gopls
            pkgs.parallel
            pkgs.git
            pkgs.cacert
            freeCodeCli
          ];
          shellHook = ''
            for skill in skills/*/SKILL.md; do
              name=$(basename $(dirname $skill))
              mkdir -p ".claude/skills/$name"
              ln -sfn "$(pwd)/$skill" ".claude/skills/$name/SKILL.md"
            done
            # The runner invokes the agent as `claude` by default; if free-code
            # ships its binary under a different name, export AUTOPATCHER_CLI to
            # match (see internal/runner).
            echo "auto-patcher dev shell"
            echo "  build: go build -o auto-patcher ."
            echo "  scan:  ./auto-patcher scan --org auto-patcher --exclude skills"
          '';
        };

        formatter = pkgs.nixfmt;
      }
    );
}
