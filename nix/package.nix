{
  lib,
  goEnv,
  buildGo127Module,
  rev ? "unknown",
}:

let
  # Last release, kept by release-please (#644). Untagged builds report it
  # too; the commit (rev) tells them apart.
  version = lib.trim (builtins.readFile ../version.txt);
  args = {
    pname = "jobshop";
    inherit version;

    src = lib.fileset.toSource {
      root = ../.;
      fileset = lib.fileset.unions [
        ../go.mod
        ../go.sum
        ../cmd
        ../pkg
      ];
    };

    ldflags = [
      "-s"
      "-w"
      "-X github.com/elee1766/jobshop/pkg/cli.version=${version}"
      "-X github.com/elee1766/jobshop/pkg/cli.commit=${rev}"
    ];

    # Tests run in CI with -race; the Nix build only produces the binary.
    doCheck = false;

    meta = {
      description = "Multi-agent work orchestration";
      mainProgram = "jobshop";
      platforms = lib.platforms.linux;
    };
  };
in
# go2nix builds one derivation per Go package, so an edit rebuilds only what
# imports it. It needs its Nix plugin (packages.go2nix-plugin) loaded into the
# evaluator; without it, fall back to one buildGoModule derivation so hosts
# and users without the plugin still build.
if builtins ? resolveGoPackages then
  goEnv.buildGoApplication (
    args
    // {
      # Module hashes come from go.sum (no lockfile).
      subPackages = [ "./cmd/jobshop" ];
      # modernc.org/sqlite is pure Go.
      CGO_ENABLED = "0";
    }
  )
else
  buildGo127Module (
    args
    // {
      subPackages = [ "cmd/jobshop" ];
      # Fallback builder only. Update with: nix build .#jobshop 2>&1 | grep got:
      # (without the plugin loaded).
      vendorHash = "sha256-2S0ym1ufgWP7qKax8+RVToAmNBniDX2b3uHHAetF5Lk=";
      env.CGO_ENABLED = "0";
    }
  )
