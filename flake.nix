{
  description = "Jobshop multi-agent work orchestration";

  # Same channel gasserver runs; ships buildGo127Module (Go 1.27.3).
  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-26.05";

  outputs =
    { self, nixpkgs }:
    let
      systems = [
        "x86_64-linux"
        "aarch64-linux"
      ];
      forAllSystems = f: nixpkgs.lib.genAttrs systems (system: f nixpkgs.legacyPackages.${system});
    in
    {
      packages = forAllSystems (pkgs: rec {
        jobshop = pkgs.callPackage ./nix/package.nix {
          rev = self.shortRev or self.dirtyShortRev or "unknown";
        };
        default = jobshop;
      });

      nixosModules.jobshop = import ./nix/module.nix self;
      nixosModules.default = self.nixosModules.jobshop;

      devShells = forAllSystems (pkgs: {
        default = pkgs.mkShell {
          packages = [
            pkgs.go_1_27
            pkgs.gopls
            pkgs.golangci-lint
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
