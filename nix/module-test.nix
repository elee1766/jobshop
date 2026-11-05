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
  ident = eval {
    gitIdentity.jobshopIdentities = true;
    identities.elee-agent = {
      kind = "github-app";
      app_id = 1;
      private_key_file = "/run/secrets/k.pem";
      installations.elee1766 = 2;
      repos = [ "elee1766/*" ];
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

  # Shared read-only mise installs (#316).
  miseShared = eval {
    sandbox.enable = true;
    sandbox.mise.tools = [ "node@24" ];
  };
  vlLocal = eval { journal.victorialogs.enable = true; };
  vlExternal = eval { journal.victorialogs.url = "http://logs.internal:9428"; };
  miseOff = eval { sandbox.enable = true; };

  # User-managed jobshop.yaml (#531): the module writes only
  # jobshop.module.yaml and watches the user's file.
  userCfg = eval {
    configFile = "/etc/jobshop/jobshop.yaml";
    identities.elee-agent = {
      kind = "github-token";
      token_file = "/run/secrets/gh";
      repos = [ "elee1766/*" ];
    };
    sandbox.enable = true;
  };
  inPlace = eval { configFile = "/var/lib/jobshop/jobshop.yaml"; };
  # Config repo (#532): jobshop.yaml lives in a workspace checkout.
  repoCfg = eval {
    workspaces.cfg = {
      url = "https://example.com/me/jobshop-config.git";
      config = "jobshop.yaml";
    };
  };
  repoSync = pkgs.writeText "config-repo-sync" (
    lib.replaceStrings
      [
        "${pkgs.systemd}/bin/systemctl"
        "${lib.getExe pkgs.hello} reload"
        "${lib.getExe pkgs.hello}"
        "/var/lib/jobshop"
      ]
      [
        "$T/bin/systemctl"
        "$T/bin/reload"
        "${lib.getExe self.packages.${pkgs.stdenv.hostPlatform.system}.jobshop}"
        "$T/root"
      ]
      repoCfg.config.systemd.services.jobshop-workspace-cfg-sync.script
  );
  configRepoTest = ''
    export T=$PWD/r HOME=$PWD/r GIT_AUTHOR_NAME=t GIT_AUTHOR_EMAIL=t@t GIT_COMMITTER_NAME=t GIT_COMMITTER_EMAIL=t@t
    export PATH=${pkgs.git}/bin:$PATH
    mkdir -p $T/bin $T/root/workspaces
    printf '#!${pkgs.runtimeShell}\nexit 0\n' > $T/bin/systemctl
    printf '#!${pkgs.runtimeShell}\necho reload >> $T/log\n' > $T/bin/reload
    chmod +x $T/bin/*
    rfail() { echo "config repo test: $*"; cat $T/log $T/out 2>/dev/null; exit 1; }
    cp ${linkTarget repoCfg "jobshop.module.yaml"} $T/root/jobshop.module.yaml
    grep -q 'config_repo_file: /var/lib/jobshop/config-repo.failed' $T/root/jobshop.module.yaml || rfail "no attention file in module file"
    git init -q --bare -b main $T/origin.git
    git clone -q $T/origin.git $T/work 2>/dev/null
    cd $T/work
    mkdir -p packs/p
    printf 'include: [jobshop.module.yaml]\nprojects:\n  js: {}\nagents:\n  a: {command: x}\npacks:\n  - path: packs/p\n' > jobshop.yaml
    pool() { printf 'pools:\n  dev:\n    agent: a\n    dir: /w\n    prompt: hi\n    max_sessions: %s\n' "$1" > packs/p/pack.yaml; }
    push() { git add -A && git commit -qm "$1" && git push -q origin HEAD:main; }
    pool 1; push one
    git clone -q $T/origin.git $T/root/workspaces/cfg
    ln -s $T/root/workspaces/cfg/jobshop.yaml $T/root/jobshop.yaml
    cd $T
    run() { : > $T/log; rc=0; (cd $T && ${pkgs.bash}/bin/bash ${repoSync}) > $T/out 2>&1 || rc=$?; }
    cur() { git -C $T/root/workspaces/cfg rev-parse HEAD; }

    # A push that changes a pool: checked, then reloaded.
    (cd $T/work && pool 2 && push two)
    run
    [ $rc = 0 ] || rfail "good push exit $rc"
    grep -q reload $T/log || rfail "good push not reloaded"
    good=$(cur)
    [ "$(cat $T/root/config-repo.last-good)" = "$good" ] || rfail "last-good not recorded"
    grep -q 'max_sessions: 2' $T/root/workspaces/cfg/packs/p/pack.yaml || rfail "checkout not moved"

    # A broken pack: reset to last good, attention file, no reload.
    (cd $T/work && printf 'pools: [\n' > packs/p/pack.yaml && push broken)
    run
    [ $rc = 1 ] || rfail "broken push exit $rc"
    ! grep -q reload $T/log || rfail "broken push reloaded"
    [ "$(cur)" = "$good" ] || rfail "not reset to last good"
    [ -s $T/root/config-repo.failed ] || rfail "no attention file"
    grep -q pack.yaml $T/root/config-repo.failed || rfail "attention file lacks the error"
    # The same bad commit is not retried every tick.
    run
    [ $rc = 0 ] || rfail "repeat run exit $rc"
    [ "$(cur)" = "$good" ] || rfail "repeat run moved the checkout"

    # The next good push recovers.
    (cd $T/work && pool 3 && push fixed)
    run
    [ $rc = 0 ] || rfail "fix push exit $rc"
    grep -q reload $T/log || rfail "fix not reloaded"
    [ ! -e $T/root/config-repo.failed ] || rfail "attention file not cleared"
    [ "$(cat $T/root/config-repo.last-good)" = "$(cur)" ] || rfail "last-good not advanced"

    # A local edit that blocks the fast-forward raises attention.
    echo edit >> $T/root/workspaces/cfg/packs/p/pack.yaml
    git -C $T/root/workspaces/cfg commit -qam local
    (cd $T/work && pool 4 && push four)
    run
    [ $rc = 1 ] || rfail "diverged exit $rc"
    grep -q 'not fast-forwardable' $T/root/config-repo.failed || rfail "diverged checkout not in attention"
    cd ..
  '';
  rules = c: c.config.systemd.tmpfiles.rules;
  linkTarget =
    c: name:
    lib.last (lib.splitString " " (lib.findFirst (r: lib.hasInfix "/${name} " r) "" (rules c)));
  # The check-and-reload script with the real jobshop binary and stubs
  # for runuser and systemctl; the root is moved to $T.
  reloadScript = pkgs.writeText "jobshop-config" (
    lib.replaceStrings
      [
        "${pkgs.util-linux}/bin/runuser"
        "${pkgs.systemd}/bin/systemctl"
        "${lib.getExe pkgs.hello}"
        "/var/lib/jobshop/jobshop.yaml"
      ]
      [
        "$T/bin/runuser"
        "$T/bin/systemctl"
        "${lib.getExe self.packages.${pkgs.stdenv.hostPlatform.system}.jobshop}"
        "$T/root/jobshop.yaml"
      ]
      userCfg.config.systemd.services.jobshop-config.script
  );
  userConfigTest = ''
    export T=$PWD/u
    mkdir -p $T/bin $T/root
    printf '#!${pkgs.runtimeShell}\nshift 3; exec "$@"\n' > $T/bin/runuser
    printf '#!${pkgs.runtimeShell}\necho "systemctl $*" >> $T/log\n' > $T/bin/systemctl
    chmod +x $T/bin/*
    cp ${linkTarget userCfg "jobshop.module.yaml"} $T/root/jobshop.module.yaml
    ufail() { echo "user config test: $*"; cat $T/log $T/out 2>/dev/null; exit 1; }
    grep -q 'token_file: /run/secrets/gh' $T/root/jobshop.module.yaml || ufail "identity not in module file"
    grep -q 'backend: bubblewrap' $T/root/jobshop.module.yaml || ufail "sandbox not in module file"
    printf 'include: [jobshop.module.yaml]\nprojects:\n  js: {identities: [elee-agent]}\n' > $T/root/jobshop.yaml
    cp $T/root/jobshop.yaml $T/want
    : > $T/log
    ${pkgs.bash}/bin/bash ${reloadScript} > $T/out 2>&1 || ufail "good edit failed the check"
    grep -q 'systemctl try-reload-or-restart jobshop.service' $T/log || ufail "good edit not reloaded"
    cmp -s $T/root/jobshop.yaml $T/want || ufail "user file changed"
    # js names an identity only the module file defines: without the
    # include the check fails and nothing reloads.
    printf 'projects:\n  js: {identities: [elee-agent]}\n' > $T/root/jobshop.yaml
    : > $T/log
    ! ${pkgs.bash}/bin/bash ${reloadScript} > $T/out 2>&1 || ufail "broken edit passed the check"
    ! grep -q reload $T/log || ufail "broken edit reloaded"
    grep -q 'keeping the running config' $T/out || ufail "no log line for the broken edit"
  '';

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
    stub jobshop '[ "$1" = version ] && echo "jobshop v0 ($(cat $T/cur))"; if [ "$1" = config ]; then echo "check $*" >> $T/log; [ -e $T/badcheck ] && { echo "pools.operator.prompt_file: var \"jobshop_repo\" is not set"; exit 1; }; echo ok; fi; exit 0'
    stub nix 'case "$1" in flake) echo "{\"path\":\"$T/src\"}" ;; build) echo $T ;; esac'
    stub jq 'sed -n "s/.*\"path\":\"\([^\"]*\)\".*/\1/p"'
    stub systemctl 'exit 0'
    stub nixos-rebuild 'echo "rebuild $*" >> $T/log; [ -n "$(cat $T/shipbad 2>/dev/null)" ] || cat $T/green > $T/cur'
    stub nix-env 'echo "nix-env $*" >> $T/log'
    printf '#!${pkgs.runtimeShell}\necho "switch gen1 $1" >> $T/log\n' > $T/gen1/bin/switch-to-configuration
    chmod +x $T/gen1/bin/switch-to-configuration
    ln -s $T/gen1 $T/system
    export PATH=$T/bin:$PATH
    run() { : > $T/log; rc=0; ${pkgs.bash}/bin/bash ${deployScript} > $T/out 2>&1 || rc=$?; }
    fail() { echo "rollback test: $*"; cat $T/log $T/out; exit 1; }

    echo 0000000 > $T/cur
    echo abcdef1234 > $T/green

    # Config check (#261): a pack var the host config misses fails before
    # any switch, marks the rev failed and keeps the reason for attention.
    touch $T/badcheck
    echo 9999999aaa > $T/green
    run
    [ $rc = 1 ] || fail "bad check exit $rc"
    ! grep -q rebuild $T/log || fail "switched despite a failed config check"
    grep -q "check config check --file /var/lib/jobshop/jobshop.yaml --replace /var/lib/jobshop/src/jobshop=$T/src" $T/log || fail "config check not run against the new source"
    [ "$(cat $T/state/failed)" = 9999999aaa ] || fail "failed rev not recorded"
    grep -q 'jobshop_repo' $T/state/failed.reason || fail "reason not kept"
    rm $T/badcheck $T/state/failed $T/state/failed.reason
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

    # The unit runs another commit than deployed says (#225): the next run
    # rewrites deployed with a WARN and keeps its mtime.
    echo 53b6b27 > $T/cur
    echo 53b6b27 > $T/green
    touch -d @1000000000 $T/state/deployed
    run
    [ $rc = 0 ] || fail "reconcile run exit $rc"
    [ "$(cat $T/state/deployed)" = 53b6b27 ] || fail "deployed not reconciled"
    [ "$(stat -c %Y $T/state/deployed)" = 1000000000 ] || fail "deployed mtime changed"
    [ "$(grep -c 'switched outside self-deploy.*deployed rewritten' $T/out)" = 1 ] || fail "want one WARN"
    ! grep -q rebuild $T/log || fail "reconcile rebuilt"
  '';

  units = c: lib.attrNames c.config.systemd.services;
  expect = name: cond: if cond then true else throw "module-test: ${name} failed";
  checks = [
    (expect "shared mise unit, env and mount" (
      let
        svc = miseShared.config.systemd.services;
      in
      svc.jobshop.environment.MISE_SHARED_INSTALL_DIRS == "/var/lib/jobshop-mise/installs"
      && svc.jobshop-mise.environment.MISE_DATA_DIR == "/var/lib/jobshop-mise"
      && lib.hasInfix "mise install --yes node@24" svc.jobshop-mise.script
      && miseShared.config.systemd.timers ? jobshop-mise
      && !(miseOff.config.systemd.services ? jobshop-mise)
      && !(miseOff.config.systemd.services.jobshop.environment ? MISE_SHARED_INSTALL_DIRS)
    ))
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
    # #216: identities go to the yaml, git and gh get tokens from jobshop.
    (expect "jobshopIdentities wires git and gh" (
      let
        env = ident.config.systemd.services.jobshop.environment;
        vals = lib.filter (v: v != null) (lib.genList (i: env."GIT_CONFIG_VALUE_${toString i}" or null) 4);
      in
      failed ident == [ ]
      && lib.any (lib.hasSuffix "identity git-credential") vals
      && lib.elem "true" vals
      && (lib.head ident.config.systemd.services.jobshop.path).name == "gh"
    ))
    (expect "jobshopIdentities and credentialHelper conflict" (
      lib.any (lib.hasInfix "not both") (
        failed (eval {
          gitIdentity.jobshopIdentities = true;
          gitIdentity.credentialHelper = "x";
        })
      )
    ))
    (expect "desk passes the proxy user flags" (
      let
        d = eval {
          desk = {
            enable = true;
            userHeader = "Tailscale-User-Login";
            trustedProxies = [ "127.0.0.1" ];
          };
        };
        exec = d.config.systemd.services.jobshop-desk.serviceConfig.ExecStart;
      in
      failed d == [ ]
      && lib.hasInfix "--user-header Tailscale-User-Login --trusted-proxy 127.0.0.1" exec
      && !lib.hasInfix "--allow-remote" exec
    ))
    (expect "desk userHeader needs trustedProxies" (
      lib.any (lib.hasInfix "trustedProxies") (
        failed (eval {
          desk = {
            enable = true;
            userHeader = "X-Forwarded-User";
          };
        })
      )
    ))
    # #516: Desk under a path prefix behind a proxy.
    (expect "desk.basePath and trustedProxies reach jobshop desk" (
      let
        s =
          (eval {
            desk = {
              enable = true;
              basePath = "/desk/";
              trustedProxies = [ "127.0.0.1" ];
            };
          }).config.systemd.services.jobshop-desk.serviceConfig.ExecStart;
      in
      lib.hasInfix "desk --listen 127.0.0.1:7480 --base-path /desk/ --trusted-proxy 127.0.0.1" s
    ))
    (expect "configFile mode links the user file and writes only the module file" (
      failed userCfg == [ ]
      && linkTarget userCfg "jobshop.yaml" == "/etc/jobshop/jobshop.yaml"
      && lib.hasSuffix "jobshop.module.yaml" (linkTarget userCfg "jobshop.module.yaml")
      &&
        userCfg.config.systemd.paths.jobshop-config.pathConfig.PathChanged == [
          "/etc/jobshop/jobshop.yaml"
          "/var/lib/jobshop/jobshop.module.yaml"
        ]
    ))
    (expect "configFile = dataDir/jobshop.yaml leaves it alone" (
      failed inPlace == [ ] && !lib.any (lib.hasInfix "/var/lib/jobshop/jobshop.yaml ") (rules inPlace)
    ))
    (expect "managed mode has no module file or watcher" (
      !lib.any (lib.hasInfix "jobshop.module.yaml") (rules two)
      && !(two.config.systemd.paths ? jobshop-config)
    ))
    (expect "configFile and settings conflict" (
      lib.any (lib.hasInfix "configFile or settings") (
        failed (eval {
          configFile = "/etc/jobshop/jobshop.yaml";
          settings.projects.js = { };
        })
      )
    ))
    (expect "config repo links the checkout file and syncs on a timer" (
      failed repoCfg == [ ]
      && linkTarget repoCfg "jobshop.yaml" == "/var/lib/jobshop/workspaces/cfg/jobshop.yaml"
      && repoCfg.config.systemd.timers ? jobshop-workspace-cfg-sync
      &&
        repoCfg.config.systemd.paths.jobshop-config.pathConfig.PathChanged == [
          "/var/lib/jobshop/jobshop.module.yaml"
        ]
    ))
    (expect "config repo and configFile conflict" (
      lib.any (lib.hasInfix "at most one workspace") (
        failed (eval {
          configFile = "/etc/jobshop/jobshop.yaml";
          workspaces.cfg = {
            url = "https://example.com/c.git";
            config = "jobshop.yaml";
          };
        })
      )
    ))
    (expect "journal.victorialogs.enable runs a local VictoriaLogs" (
      vlLocal.config.services.victorialogs.enable
      && vlLocal.config.services.victorialogs.listenAddress == "127.0.0.1:9428"
      && lib.elem "-retentionPeriod=90d" vlLocal.config.services.victorialogs.extraOptions
    ))
    (expect "an external VictoriaLogs url runs none" (!vlExternal.config.services.victorialogs.enable))
    (expect "per_session ids past subIdCount fail, inside pass" (
      let
        iso = range: {
          isolation.enable = true;
          isolation.subIdCount = 100;
          settings.pools.dev.isolation.user = {
            uid = 1000;
            map_uid = 231072;
            map_gid = 231072;
            groups = [ "jobshop" ];
            per_session.range = range;
          };
        };
      in
      lib.any (lib.hasInfix "pool dev per_session") (failed (eval (iso 101)))
      && !lib.any (lib.hasInfix "per_session") (failed (eval (iso 100)))
    ))
  ];
in
# Assertions run at evaluation, so `nix flake check --no-build` covers them;
# the pack paths in the generated file are checked when the check builds.
assert lib.all lib.id checks;
pkgs.runCommand "jobshop-module-test" { } ''
  ${rollbackTest}
  ${userConfigTest}
  ${configRepoTest}
  grep -q '/var/lib/jobshop/workspaces/jobshop/packs/selfdev' ${yamlPacks two}
  grep -q 'private_key_file: /run/secrets/k.pem' ${yamlPacks ident}
  grep -q 'exec .* gh "\$@"' ${lib.head ident.config.systemd.services.jobshop.path}/bin/gh
  grep -q '/var/lib/jobshop/workspaces/other/packs/github' ${yamlPacks two}
  grep -q 'repo: elee1766/other' ${yamlPacks two}
  grep -q 'prefix: ot-' ${yamlPacks two}
  [ "$(grep -c 'prefix:' ${yamlPacks two})" = 1 ] || { echo 'selfdev got a prefix'; exit 1; }
  grep -q -- '- /var/lib/jobshop-mise' ${yamlPacks miseShared}
  grep -q 'url: http://127.0.0.1:9428' ${yamlPacks vlLocal}
  grep -q 'url: http://logs.internal:9428' ${yamlPacks vlExternal}
  touch $out
''
