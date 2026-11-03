# Evaluation checks for the NixOS module (no VM): `nix flake check` builds
# this, so a failed expectation fails evaluation.
{
  self,
  nixpkgs,
  pkgs,
}:
let
  lib = nixpkgs.lib;
  eval =
    jobshop:
    nixpkgs.lib.nixosSystem {
      inherit (pkgs.stdenv.hostPlatform) system;
      modules = [
        self.nixosModules.jobshop
        {
          system.stateVersion = "26.05";
          boot.loader.grub.enable = false;
          fileSystems."/" = {
            device = "none";
            fsType = "tmpfs";
          };
          services.jobshop = {
            enable = true;
            package = pkgs.hello;
          }
          // jobshop;
        }
      ];
    };
  failed = c: map (a: a.message) (lib.filter (a: !a.assertion) c.config.assertions);
  # The generated jobshop.yaml, from its tmpfiles link rule.  Live packs are
  # added there, not to the settings option.
  yamlPacks =
    c:
    let
      link = lib.findFirst (r: lib.hasInfix "/jobshop.yaml " r) "" c.config.systemd.tmpfiles.rules;
    in
    lib.last (lib.splitString " " link);

  two = eval {
    workspaces.jobshop = {
      url = "https://github.com/elee1766/jobshop.git";
      livePacks = [ "packs/selfdev" ];
      selfDeploy = true;
    };
    workspaces.other = {
      url = "git@github.com:elee1766/other.git";
      livePacks = [ "packs/github" ];
      livePackVars.repo = "elee1766/other";
      livePackPrefix = "ot-";
    };
  };
  bothDeploy = eval {
    workspaces.a = {
      url = "https://github.com/x/a.git";
      selfDeploy = true;
    };
    workspaces.b = {
      url = "https://github.com/x/b.git";
      selfDeploy = true;
    };
  };
  # The old option, as gasserver sets it.
  legacy = eval {
    workspace = {
      url = "https://github.com/elee1766/jobshop.git";
      livePacks = [ "packs/selfdev" ];
    };
    selfDeploy = {
      enable = true;
      hostFlake = "x#y";
      inputUrl = "git+https://example/jobshop?ref=master";
    };
  };

  # The self-deploy script run against stubs: host paths are rewritten to a
  # temp dir, gh/nixos-rebuild/nix-env/systemctl and the controller binary are
  # shell stubs that log their calls.
  deployScript = pkgs.writeText "self-deploy" (
    lib.replaceStrings
      [
        "state=/var/lib/jobshop-deploy"
        "${pkgs.util-linux}/bin/runuser"
        "gh run list"
        ''readlink -f "/proc/$(systemctl show -p MainPID --value jobshop.service)/exe"''
        "/nix/var/nix/profiles/system"
        "sleep 2"
      ]
      [
        "state=$T/state"
        "runuser"
        "$T/bin/gh run list"
        "echo $T/bin/jobshop"
        "$T/system"
        ":"
      ]
      legacy.config.systemd.services.jobshop-self-deploy.script
  );
  rollbackTest = ''
    export T=$PWD/t
    mkdir -p $T/bin $T/state $T/gen1/bin
    stub() { printf '#!${pkgs.runtimeShell}\n%s\n' "$2" > $T/bin/$1; chmod +x $T/bin/$1; }
    stub runuser 'shift 3; exec "$@"'
    stub gh 'cat $T/green'
    stub jobshop '[ "$1" = version ] && echo "jobshop v0 ($(cat $T/cur))"; exit 0'
    stub systemctl 'exit 0'
    stub nixos-rebuild 'echo "rebuild $*" >> $T/log; [ -n "$(cat $T/shipbad 2>/dev/null)" ] || cat $T/green > $T/cur'
    stub nix-env 'echo "nix-env $*" >> $T/log'
    printf '#!${pkgs.runtimeShell}\necho "switch gen1 $1" >> $T/log\n' > $T/gen1/bin/switch-to-configuration
    chmod +x $T/gen1/bin/switch-to-configuration
    ln -s $T/gen1 $T/system
    export PATH=$T/bin:$PATH
    run() { : > $T/log; rc=0; ${pkgs.bash}/bin/bash ${deployScript} || rc=$?; }
    fail() { echo "rollback test: $*"; cat $T/log; exit 1; }

    echo 0000000 > $T/cur
    echo abcdef1234 > $T/green

    # Drill: the switch succeeds, the drill marks it unhealthy.
    touch $T/state/drill
    run
    [ $rc = 1 ] || fail "drill run exit $rc"
    [ ! -e $T/state/drill ] || fail "drill marker not cleared"
    [ "$(cat $T/state/failed)" = abcdef1234 ] || fail "rev not written to failed"
    grep -q "nix-env -p $T/system --set $T/gen1" $T/log || fail "no rollback to prev"
    grep -q "switch gen1 switch" $T/log || fail "prev not switched to"
    [ ! -e $T/state/deployed ] || fail "deployed written on rollback"

    # The same rev is not retried.
    echo 0000000 > $T/cur
    run
    [ $rc = 0 ] || fail "retry run exit $rc"
    ! grep -q rebuild $T/log || fail "failed rev rebuilt again"

    # Without the drill, a controller still on the old binary is unhealthy too.
    echo bbbbbbb999 > $T/green
    echo 1 > $T/shipbad
    run
    [ $rc = 1 ] || fail "unhealthy run exit $rc"
    [ "$(cat $T/state/failed)" = bbbbbbb999 ] || fail "unhealthy rev not in failed"

    # A newer healthy rev deploys.
    rm $T/shipbad
    echo ccccccc777 > $T/green
    run
    [ $rc = 0 ] || fail "healthy run exit $rc"
    [ "$(cat $T/state/deployed)" = ccccccc777 ] || fail "deployed not written"
  '';

  units = c: lib.attrNames c.config.systemd.services;
  expect = name: cond: if cond then true else throw "module-test: ${name} failed";
  checks = [
    (expect "two workspaces pass assertions" (failed two == [ ]))
    (expect "two workspaces have clone units" (
      lib.all (u: lib.elem u (units two)) [
        "jobshop-workspace-jobshop"
        "jobshop-workspace-other"
        "jobshop-workspace-other-sync"
      ]
    ))
    (expect "two selfDeploy workspaces fail" (
      lib.any (lib.hasInfix "more than one workspace") (failed bothDeploy)
    ))
    (expect "legacy workspace maps to workspaces.default" (
      legacy.config.services.jobshop.workspaces.default.dir == "/var/lib/jobshop/src/jobshop"
      && legacy.config.services.jobshop.workspaces.default.selfDeploy
      && failed legacy == [ ]
    ))
    (expect "self-deploy uses the selfDeploy workspace's repo" (
      lib.hasInfix "--repo elee1766/jobshop " legacy.config.systemd.services.jobshop-self-deploy.script
    ))
    # A live pack extends packs outside livePacks, so the sync reloads on
    # any checkout move, not on a diff of the live paths only (#177).
    (expect "sync reloads on any checkout move" (
      let
        s = legacy.config.systemd.services.jobshop-workspace-default-sync.script;
      in
      lib.hasInfix ''if [ "$before" != "$after" ]; then'' s && !(lib.hasInfix "git diff --quiet" s)
    ))
    # A yaml-only change must reload the controller (#210): the unit's
    # X-Reload-Triggers follow the generated file, and the binary does not.
    (expect "yaml-only change reloads the controller" (
      let
        trig = c: c.config.systemd.services.jobshop.reloadTriggers;
        start = c: c.config.systemd.services.jobshop.serviceConfig.ExecStart;
        more = eval {
          workspaces.jobshop = {
            url = "https://github.com/elee1766/jobshop.git";
            livePacks = [ "packs/selfdev" ];
            livePackVars.dev_sessions = 8;
          };
        };
        base = eval {
          workspaces.jobshop = {
            url = "https://github.com/elee1766/jobshop.git";
            livePacks = [ "packs/selfdev" ];
            livePackVars.dev_sessions = 4;
          };
        };
      in
      map toString (trig base) == [ (yamlPacks base) ]
      && map toString (trig base) != map toString (trig more)
      && start base == start more
    ))
    # The sync never moves the checkout past the installed binary, even
    # with the controller down: the commit comes from the package, not
    # /proc/<MainPID>, and an unknown commit stops the sync (#227).
    (expect "sync follows the installed binary" (
      let
        s = builtins.unsafeDiscardStringContext legacy.config.systemd.services.jobshop-workspace-default-sync.script;
      in
      lib.hasInfix "installed=$(${builtins.unsafeDiscardStringContext (lib.getExe legacy.config.services.jobshop.package)} version" s
      && lib.hasInfix "target=$installed" s
      && lib.hasInfix "not moving it" s
      && !(lib.hasInfix "MainPID" s)
    ))
    (expect "self-deploy logs a switch outside it" (
      lib.hasInfix "switched outside self-deploy" legacy.config.systemd.services.jobshop-self-deploy.script
    ))
  ];
in
# Assertions run at evaluation, so `nix flake check --no-build` covers them;
# the pack paths in the generated file are checked when the check builds.
assert lib.all lib.id checks;
pkgs.runCommand "jobshop-module-test" { } ''
  ${rollbackTest}
  grep -q '/var/lib/jobshop/workspaces/jobshop/packs/selfdev' ${yamlPacks two}
  grep -q '/var/lib/jobshop/workspaces/other/packs/github' ${yamlPacks two}
  grep -q 'repo: elee1766/other' ${yamlPacks two}
  grep -q 'prefix: ot-' ${yamlPacks two}
  [ "$(grep -c 'prefix:' ${yamlPacks two})" = 1 ] || { echo 'selfdev got a prefix'; exit 1; }
  touch $out
''
