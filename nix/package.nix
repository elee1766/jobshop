{
  lib,
  goEnv,
  buildGo127Module,
  runCommand,
  deskUi,
  rev ? "unknown",
}:

let
  # Last release, kept by release-please (#644). It stays the derivation
  # version so store paths do not change per commit.
  version = lib.trim (builtins.readFile ../version.txt);
  # The binary reports the commit as semver build metadata (JOB-79):
  # self-deploy builds master commits after a release, which would
  # otherwise report that release and look one version behind.
  binVersion = if rev == "unknown" then version else "${version}+${rev}";
  goSource = lib.fileset.toSource {
    root = ../.;
    fileset = lib.fileset.unions [
      ../go.mod
      ../go.sum
      ../cmd
      (lib.fileset.difference ../pkg (lib.fileset.maybeMissing ../pkg/transport/desk/dist))
    ];
  };
  args = {
    pname = "jobshop";
    inherit version;

    src = goSource;

    ldflags = [
      "-s"
      "-w"
      "-X github.com/elee1766/jobshop/pkg/transport/cli.version=${binVersion}"
      "-X github.com/elee1766/jobshop/pkg/transport/cli.commit=${rev}"
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
      # Declare generated embeds without building the frontend during evaluation.
      packageOverrides."github.com/elee1766/jobshop/pkg/transport/desk".srcOverlay =
        runCommand "jobshop-desk-source" { }
          ''
            mkdir -p "$out"
            cp -r ${deskUi} "$out/dist"
          '';
      # modernc.org/sqlite is pure Go.
      CGO_ENABLED = "0";
    }
  )
else
  buildGo127Module (
    args
    // {
      subPackages = [ "cmd/jobshop" ];
      postPatch = ''
        cp -r ${deskUi}/. pkg/transport/desk/dist/
      '';
      # Fallback builder only. Update with: nix build .#jobshop 2>&1 | grep got:
      # (without the plugin loaded).
      vendorHash = "sha256-d7kMgHHI5D45O9igvZSzM6+7Kvzi2Mkj4dI8YNA1T7c=";
      env.CGO_ENABLED = "0";
    }
  )
