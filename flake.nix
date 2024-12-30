{
  description = "AI Toolbox - Collection of AI-related CLI tools";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-parts.url = "github:hercules-ci/flake-parts";
    git-hooks-nix.url = "github:cachix/git-hooks.nix";
  };

  outputs = inputs @ {flake-parts, ...}:
    flake-parts.lib.mkFlake {inherit inputs;} {
      imports = [
        inputs.git-hooks-nix.flakeModule
      ];
      systems = ["x86_64-linux" "aarch64-linux" "aarch64-darwin" "x86_64-darwin"];

      perSystem = {
        config,
        self',
        pkgs,
        ...
      }: let
        buildTool = name:
          pkgs.buildGoModule {
            pname = name;
            version = "0.1.0";
            src = ./tools/${name};
            vendorHash = null;
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
          golangci-lint.enable = true;
          govet.enable = true;
          deadnix.enable = true;
          trim-trailing-whitespace.enable = true;
        };

        formatter = pkgs.alejandra;
      };
    };
}
