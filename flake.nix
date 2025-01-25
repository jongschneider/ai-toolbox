{
  description = "AI Toolbox - Collection of AI-related CLI tools";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-parts.url = "github:hercules-ci/flake-parts";
    git-hooks-nix.url = "github:cachix/git-hooks.nix";
  };

  outputs = inputs @ {
    flake-parts,
    git-hooks-nix,
    ...
  }:
    flake-parts.lib.mkFlake {inherit inputs;} {
      imports = [
        git-hooks-nix.flakeModule
      ];
      systems = ["x86_64-linux" "aarch64-linux" "aarch64-darwin" "x86_64-darwin"];

      perSystem = {
        config,
        self',
        pkgs,
        ...
      }: let
        # version = "latest";
        buildTool = name:
          pkgs.buildGoModule {
            inherit name;
            src = ./.;
            subPackages = ["tools/${name}"];
            # First try with vendorHash = null to get the correct hash
            vendorHash = "sha256-wWEas0cdoc3m7gtuCSJeW01WrXfsuKQEA+8qSjupmh8="; # update whenever go.mod changes
            # vendorHash = ""; # update whenever go.mod changes. don't foget to `ga .`  🤡
          };
      in {
        # Packages
        packages = {
          appender = buildTool "appender";
          default = self'.packages.appender;
        };

        # Apps
        apps = {
          appender = {
            type = "app";
            program = "${self'.packages.appender}/bin/appender";
          };
          default = self'.apps.appender;
        };

        # Development shell
        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            go
            gopls
            golangci-lint
            just
            delve
          ];
          shellHook = config.pre-commit.installationScript;
        };
        pre-commit.settings.hooks = {
          alejandra.enable = true;
          gofmt.enable = true;
          golines.enable = true;
          golangci-lint.enable = true;
          govet.enable = true;
          staticcheck.enable = true;
          deadnix.enable = true;
          statix.enable = true;
          trim-trailing-whitespace.enable = true;
          check-yaml.enable = true;
          yamlfmt.enable = true;
          fix-byte-order-marker.enable = true;
          flake-checker.enable = true;
          prettier.enable = true;
          ripsecrets.enable = true;
          shellcheck.enable = true;
        };

        formatter = pkgs.alejandra;
      };
    };
}
