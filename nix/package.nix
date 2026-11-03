{
  lib,
  buildGo127Module,
  rev ? "unknown",
}:

buildGo127Module {
  pname = "jobshop";
  version = "0-${rev}";

  src = lib.fileset.toSource {
    root = ../.;
    fileset = lib.fileset.unions [
      ../go.mod
      ../go.sum
      ../cmd
      ../pkg
    ];
  };

  subPackages = [ "cmd/jobshop" ];

  # Update with: nix build .#jobshop 2>&1 | grep got:
  vendorHash = "sha256-P9k4HYajre7ZE7soC6+qurP/hGm9kt1uGCqwSVf9URs=";

  # modernc.org/sqlite is pure Go.
  env.CGO_ENABLED = "0";

  ldflags = [
    "-s"
    "-w"
    "-X github.com/elee1766/jobshop/pkg/cli.version=0-${rev}"
    "-X github.com/elee1766/jobshop/pkg/cli.commit=${rev}"
  ];

  # Tests run in CI with -race; the Nix build only produces the binary.
  doCheck = false;

  meta = {
    description = "Multi-agent work orchestration";
    mainProgram = "jobshop";
    platforms = lib.platforms.linux;
  };
}
