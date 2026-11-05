{
  description = "Jobshop multi-agent work orchestration";

  # Same channel gasserver runs; ships buildGo127Module (Go 1.27.3).
  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-26.05";
  inputs.go2nix = {
    url = "github:numtide/go2nix";
    inputs.nixpkgs.follows = "nixpkgs";
  };

  outputs =
    {
      self,
      nixpkgs,
      go2nix,
    }:
    let
      systems = [
        "x86_64-linux"
        "aarch64-linux"
      ];
      forAllSystems = f: nixpkgs.lib.genAttrs systems (system: f nixpkgs.legacyPackages.${system});
    in
    {
      packages = forAllSystems (
        pkgs:
        let
          system = pkgs.stdenv.hostPlatform.system;
        in
        rec {
          jobshop = pkgs.callPackage ./nix/package.nix {
            rev = self.shortRev or self.dirtyShortRev or "unknown";
            goEnv = go2nix.lib.mkGoEnv {
              go = pkgs.go_1_27;
              inherit (pkgs) callPackage;
              go2nix = go2nix.packages.${system}.go2nix;
            };
          };
          default = jobshop;
          # Load into the evaluator (plugin-files) to build jobshop. It runs
          # `go list` with the Go it was built with, so it must be go.mod's.
          go2nix-plugin = go2nix.packages.${system}.go2nix-nix-plugin.override {
            pkgs = pkgs // {
              go = pkgs.go_1_27;
            };
          };
        }
      );

      nixosModules.jobshop = import ./nix/module.nix self;
      nixosModules.default = self.nixosModules.jobshop;

      devShells = forAllSystems (pkgs: {
        default = pkgs.mkShell {
          packages = [
            pkgs.go_1_27
            pkgs.gopls
            pkgs.sqlite
          ];
        };
      });

      checks = forAllSystems (pkgs: {
        module = import ./nix/module-test.nix { inherit self nixpkgs pkgs; };
      });

      formatter = forAllSystems (pkgs: pkgs.nixfmt);
    };
}
