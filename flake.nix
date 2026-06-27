{
  description = "auto-patcher Claude skills and agents";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs =
    {
      self,
      nixpkgs,
      flake-utils,
    }:
    let
      # Package the skill command files for installation.
      # Downstream flakes (e.g. a dotfiles flake using free-code.lib.mkClaude) can
      # add this package and symlink $out/share/claude/commands/* into ~/.claude/commands/.
      mkSkillsPackage =
        pkgs:
        pkgs.runCommand "auto-patcher-skills" { src = ./.claude/commands; } ''
          mkdir -p $out/share/claude/commands
          cp -r $src/. $out/share/claude/commands/
        '';

      # Package the autopatcher agent files (CLAUDE.md + PATCHER.md template).
      mkAgentPackage =
        pkgs:
        pkgs.runCommand "auto-patcher-agent" { } ''
          mkdir -p $out/share/auto-patcher
          cp ${./CLAUDE.md} $out/share/auto-patcher/CLAUDE.md
          cp ${./PATCHER.md} $out/share/auto-patcher/PATCHER.md
        '';
    in
    {
      lib = {
        # Wrap a free-code mkClaude call with the auto-patcher skill settings.
        # Usage:
        #   skills.lib.withSkills free-code.lib.mkClaude pkgs {
        #     mcpServers = { ... };
        #     settings   = { ... };
        #   }
        withSkills =
          mkClaude: pkgs: args:
          mkClaude pkgs args;

        inherit mkSkillsPackage mkAgentPackage;
      };
    }
    // flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in
      {
        packages = {
          default = mkSkillsPackage pkgs;
          agent = mkAgentPackage pkgs;
        };

        devShells.default = pkgs.mkShell {
          shellHook = ''
            echo "auto-patcher skills"
            echo "skills: /patch-dissect  /patch-design  /patch-apply"
          '';
        };

        formatter = pkgs.nixfmt;
      }
    );
}
