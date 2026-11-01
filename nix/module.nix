# NixOS module for running the Jobshop controller as its own user and slice.
# Host-specific values (limits, secrets, agent tooling) are set by the importing host.
self:
{
  config,
  lib,
  pkgs,
  ...
}:
let
  cfg = config.services.jobshop;
  settingsFormat = pkgs.formats.yaml { };
  tel = cfg.telemetry;
  # Only set keys go into the file, so settings.telemetry can still be used directly.
  telemetrySettings = lib.filterAttrs (_: v: v != "" && v != { }) {
    metrics_listen = tel.metricsListen;
    otlp_endpoint = tel.otlpEndpoint;
    otlp_traces_endpoint = tel.otlpTracesEndpoint;
    otlp_logs_endpoint = tel.otlpLogsEndpoint;
    otlp_headers = tel.otlpHeaders;
  };
  livePacks = map (p: {
    path = "${cfg.workspace.dir}/${p}";
    vars = cfg.workspace.livePackVars;
  }) cfg.workspace.livePacks;
  settings =
    cfg.settings
    // lib.optionalAttrs (livePacks != [ ]) { packs = (cfg.settings.packs or [ ]) ++ livePacks; };
  configFile = settingsFormat.generate "jobshop.yaml" (
    lib.recursiveUpdate settings (
      lib.optionalAttrs (telemetrySettings != { }) { telemetry = telemetrySettings; }
    )
  );
  jobshop = lib.getExe cfg.package;

  defaultTools = with pkgs; [
    git
    go_1_27
    nix
    gnumake
    gcc
    ripgrep
    fd
    jq
    curl
    openssh
    coreutils
    bash
  ];
  path = cfg.path ++ lib.optionals cfg.defaultPath defaultTools;

  # git reads GIT_CONFIG_KEY_n/VALUE_n, so this needs no gitconfig file.
  gitConfig = lib.optionals (cfg.gitIdentity.credentialHelper != "") [
    [
      "credential.https://github.com.helper"
      cfg.gitIdentity.credentialHelper
    ]
    [
      "url.https://github.com/.insteadOf"
      "git@github.com:"
    ]
  ];
  gitEnv =
    lib.optionalAttrs (gitConfig != [ ]) (
      {
        GIT_CONFIG_COUNT = toString (lib.length gitConfig);
      }
      // lib.listToAttrs (
        lib.concatLists (
          lib.imap0 (i: kv: [
            (lib.nameValuePair "GIT_CONFIG_KEY_${toString i}" (lib.elemAt kv 0))
            (lib.nameValuePair "GIT_CONFIG_VALUE_${toString i}" (lib.elemAt kv 1))
          ]) gitConfig
        )
      )
    )
    // lib.optionalAttrs (cfg.gitIdentity.name != "") {
      GIT_AUTHOR_NAME = cfg.gitIdentity.name;
      GIT_COMMITTER_NAME = cfg.gitIdentity.name;
    }
    // lib.optionalAttrs (cfg.gitIdentity.email != "") {
      GIT_AUTHOR_EMAIL = cfg.gitIdentity.email;
      GIT_COMMITTER_EMAIL = cfg.gitIdentity.email;
    };
  environment =
    gitEnv
    // cfg.environment
    // {
      JOBSHOP_ROOT = cfg.dataDir;
      HOME = cfg.dataDir;
    };

  ws = cfg.workspace;
  sd = cfg.selfDeploy;
  # GitHub owner/repo from the workspace URL, for `gh` in the deploy pick.
  ghRepo = lib.removeSuffix ".git" (
    lib.last (
      lib.splitString "github.com/" (lib.replaceStrings [ "github.com:" ] [ "github.com/" ] ws.url)
    )
  );
  deployState = "/var/lib/jobshop-deploy";
in
{
  options.services.jobshop = {
    enable = lib.mkEnableOption "the Jobshop controller";

    package = lib.mkOption {
      type = lib.types.package;
      default = self.packages.${pkgs.stdenv.hostPlatform.system}.jobshop;
      description = "Jobshop package. Pinning this store path is the deployed version.";
    };

    user = lib.mkOption {
      type = lib.types.str;
      default = "jobshop";
    };

    group = lib.mkOption {
      type = lib.types.str;
      default = "jobshop";
    };

    dataDir = lib.mkOption {
      type = lib.types.path;
      default = "/var/lib/jobshop";
      description = "Jobshop root directory: jobshop.yaml (linked from settings), database, sockets, session directories and journals.";
    };

    settings = lib.mkOption {
      type = settingsFormat.type;
      default = { };
      description = "Contents of jobshop.yaml.";
    };

    path = lib.mkOption {
      type = lib.types.listOf lib.types.package;
      default = [ ];
      description = "Tools on PATH for agents (git, gh, jcode, go, ...).";
    };

    environment = lib.mkOption {
      type = lib.types.attrsOf lib.types.str;
      default = { };
      description = "Extra environment for the controller, inherited by session hosts and agents.";
    };

    environmentFile = lib.mkOption {
      type = lib.types.nullOr lib.types.path;
      default = null;
      description = "systemd EnvironmentFile for secrets, e.g. tokens referenced as \${VAR} in telemetry.otlpHeaders. Kept out of the Nix store. Like environment, it is inherited by session hosts and agents.";
    };

    telemetry = {
      metricsListen = lib.mkOption {
        type = lib.types.str;
        default = "";
        example = "127.0.0.1:9464";
        description = "TCP address for Prometheus /metrics. Empty: off.";
      };
      otlpEndpoint = lib.mkOption {
        type = lib.types.str;
        default = "";
        example = "http://127.0.0.1:4318";
        description = "OTLP/HTTP base URL; traces go to /v1/traces and logs to /v1/logs.";
      };
      otlpTracesEndpoint = lib.mkOption {
        type = lib.types.str;
        default = "";
        example = "http://127.0.0.1:10428/insert/opentelemetry/v1/traces";
        description = "Full traces URL, overriding otlpEndpoint.";
      };
      otlpLogsEndpoint = lib.mkOption {
        type = lib.types.str;
        default = "";
        example = "http://127.0.0.1:9428/insert/opentelemetry/v1/logs";
        description = "Full logs URL, overriding otlpEndpoint.";
      };
      otlpHeaders = lib.mkOption {
        type = lib.types.attrsOf lib.types.str;
        default = { };
        example = {
          Authorization = "Bearer \${OTLP_TOKEN}";
        };
        description = "Headers on every OTLP export. \${VAR} expands from the controller environment (see environmentFile).";
      };
    };

    cpuQuota = lib.mkOption {
      type = lib.types.str;
      default = "600%";
      description = "CPUQuota for jobshop.slice (six cores' worth by default).";
    };

    memoryMax = lib.mkOption {
      type = lib.types.str;
      default = "8G";
    };

    memoryHigh = lib.mkOption {
      type = lib.types.str;
      default = "7G";
    };

    restrictEgress = lib.mkOption {
      type = lib.types.bool;
      default = true;
      description = "Deny private, CGNAT/tailnet, link-local and ULA ranges, like gasserver's agent slice.";
    };

    defaultPath = lib.mkOption {
      type = lib.types.bool;
      default = true;
      description = "Put the tools Jobshop's own agents need on PATH (git, go, nix, gnumake, gcc, ripgrep, fd, jq, curl, openssh, coreutils, bash), before `path`.";
    };

    gitIdentity = {
      name = lib.mkOption {
        type = lib.types.str;
        default = "";
        example = "elee-agent[bot]";
        description = "Author and committer name for agent commits. Empty: unset.";
      };
      email = lib.mkOption {
        type = lib.types.str;
        default = "";
      };
      credentialHelper = lib.mkOption {
        type = lib.types.str;
        default = "";
        example = "jobshop";
        description = "git credential helper for https://github.com (`git-credential-<name>` on PATH). Also rewrites git@github.com: URLs to https.";
      };
    };

    workspace = {
      url = lib.mkOption {
        type = lib.types.str;
        default = "";
        example = "https://github.com/elee1766/jobshop.git";
        description = "Repo cloned once into workspace.dir before the controller starts; pools use it as `repo`. Empty: no clone.";
      };
      dir = lib.mkOption {
        type = lib.types.str;
        default = "${cfg.dataDir}/src/${lib.removeSuffix ".git" (baseNameOf cfg.workspace.url)}";
        defaultText = lib.literalExpression ''"''${dataDir}/src/<repo name>"'';
      };
      sshCommand = lib.mkOption {
        type = lib.types.str;
        default = "";
        description = "GIT_SSH_COMMAND for the first clone only (e.g. a deploy key), when url is ssh.";
      };
      livePacks = lib.mkOption {
        type = lib.types.listOf lib.types.str;
        default = [ ];
        example = [ "packs/selfdev" ];
        description = ''
          Packs read from the live checkout (workspace.dir) instead of the Nix store, so
          merged prompt and pool changes apply after `jobshop-workspace-sync` without a
          deploy. Their entries are added to settings.packs; vars come from livePackVars.
        '';
      };
      livePackVars = lib.mkOption {
        type = lib.types.attrsOf lib.types.anything;
        default = { };
      };
      syncInterval = lib.mkOption {
        type = lib.types.str;
        default = "2min";
        description = "How often jobshop-workspace-sync fast-forwards the checkout to origin and reloads the controller if packs changed.";
      };
    };

    selfDeploy = {
      enable = lib.mkEnableOption ''
        root unit that rebuilds this host with the jobshop flake input overridden to the
        newest green commit on the workspace's default branch, then checks the controller
        and rolls back on failure. The host flake's lock is not edited'';
      hostFlake = lib.mkOption {
        type = lib.types.str;
        example = "git+ssh://git@github.com/elee1766/gasserver.nix#gasserver";
        description = "Flake output to rebuild (same as system.autoUpgrade.flake plus #host).";
      };
      inputName = lib.mkOption {
        type = lib.types.str;
        default = "jobshop";
        description = "Name of the jobshop input in the host flake.";
      };
      inputUrl = lib.mkOption {
        type = lib.types.str;
        example = "git+ssh://git@github-jobshop/elee1766/jobshop?ref=master";
        description = "URL root can fetch the jobshop flake from; `&rev=<sha>` is appended.";
      };
      branch = lib.mkOption {
        type = lib.types.str;
        default = "master";
      };
      interval = lib.mkOption {
        type = lib.types.str;
        default = "10min";
      };
    };
  };

  config = lib.mkIf cfg.enable {
    users.users.${cfg.user} = {
      isSystemUser = true;
      group = cfg.group;
      home = cfg.dataDir;
      createHome = true;
    };
    users.groups.${cfg.group} = { };

    systemd.tmpfiles.rules = [
      "d ${cfg.dataDir} 0700 ${cfg.user} ${cfg.group} -"
      # The root's config is the generated file; L+ replaces the link on each switch.
      "L+ ${cfg.dataDir}/jobshop.yaml - - - - ${configFile}"
      # Per-host GC roots keep a running session host's binary alive across Nix switches.
      "d /nix/var/nix/gcroots/per-user/${cfg.user} 0755 ${cfg.user} ${cfg.group} -"
    ];

    systemd.slices.jobshop = {
      description = "Jobshop controller, session hosts and agents";
      sliceConfig = {
        CPUQuota = cfg.cpuQuota;
        MemoryHigh = cfg.memoryHigh;
        MemoryMax = cfg.memoryMax;
        TasksMax = 8192;
        ManagedOOMMemoryPressure = "kill";
        ManagedOOMMemoryPressureLimit = "80%";
      }
      // lib.optionalAttrs cfg.restrictEgress {
        IPAddressDeny = [
          "10.0.0.0/8"
          "172.16.0.0/12"
          "192.168.0.0/16"
          "100.64.0.0/10"
          "169.254.0.0/16"
          "fc00::/7"
          "fe80::/10"
        ];
      };
    };

    systemd.services.jobshop = {
      description = "Jobshop controller";
      wantedBy = [ "multi-user.target" ];
      wants = [ "network-online.target" ];
      after = [ "network-online.target" ];
      inherit path environment;
      serviceConfig = {
        User = cfg.user;
        Group = cfg.group;
        Slice = "jobshop.slice";
        WorkingDirectory = cfg.dataDir;
        ExecStart = "${jobshop} serve";
        EnvironmentFile = lib.mkIf (cfg.environmentFile != null) cfg.environmentFile;
        # `systemctl reload jobshop` re-reads jobshop.yaml and packs.
        ExecReload = "${pkgs.coreutils}/bin/kill -HUP $MAINPID";
        # Restarts and Nix switches restart only the controller; session hosts
        # keep running and are re-adopted. Full stops go through `jobshop shutdown`.
        KillMode = "process";
        Restart = "on-failure";
        RestartSec = 2;
      };
    };

    # Full stop at system shutdown: end every session and reap hosts. Ordered
    # after jobshop.service so it stops first, while the controller is still up.
    systemd.services.jobshop-shutdown = {
      description = "Stop all Jobshop sessions at system shutdown";
      wantedBy = [ "multi-user.target" ];
      after = [ "jobshop.service" ];
      # A Nix switch must never run ExecStop here; that would end every session.
      restartIfChanged = false;
      stopIfChanged = false;
      environment.JOBSHOP_ROOT = cfg.dataDir;
      serviceConfig = {
        Type = "oneshot";
        RemainAfterExit = true;
        User = cfg.user;
        Group = cfg.group;
        Slice = "jobshop.slice";
        ExecStart = "${pkgs.coreutils}/bin/true";
        ExecStop = "${jobshop} shutdown";
        TimeoutStopSec = 120;
      };
    };

    # Clone the workspace once. Fetches and pushes later go over https with
    # the credential helper (the controller's env); sshCommand is for this
    # first clone only.
    systemd.services.jobshop-workspace = lib.mkIf (ws.url != "") {
      description = "Clone the Jobshop workspace repo";
      wantedBy = [ "jobshop.service" ];
      before = [ "jobshop.service" ];
      after = [ "network-online.target" ];
      wants = [ "network-online.target" ];
      inherit path environment;
      serviceConfig = {
        Type = "oneshot";
        RemainAfterExit = true;
        User = cfg.user;
        Group = cfg.group;
      };
      script = ''
        if [ ! -d ${ws.dir}/.git ]; then
          mkdir -p "$(dirname ${ws.dir})"
          ${
            lib.optionalString (ws.sshCommand != "") "GIT_SSH_COMMAND=${lib.escapeShellArg ws.sshCommand}"
          } \
            git clone ${lib.escapeShellArg ws.url} ${ws.dir}
        fi
        ${lib.optionalString (lib.hasPrefix "git@github.com:" ws.url) ''
          git -C ${ws.dir} remote set-url origin https://github.com/${ghRepo}.git
        ''}
      '';
    };

    # Keep the main checkout at origin so live packs follow merges; reload the
    # controller when a pack file changed. Session worktrees are separate.
    systemd.services.jobshop-workspace-sync = lib.mkIf (ws.url != "") {
      description = "Fast-forward the Jobshop workspace and reload live packs";
      after = [ "jobshop-workspace.service" ];
      requires = [ "jobshop-workspace.service" ];
      inherit path environment;
      serviceConfig = {
        Type = "oneshot";
        User = cfg.user;
        Group = cfg.group;
      };
      script = ''
        set -eu
        cd ${ws.dir}
        git fetch --quiet origin
        branch=$(git symbolic-ref --short HEAD)
        target="origin/$branch"
        ${lib.optionalString sd.enable ''
          # Packs follow the deployed binary, so a pack that needs newer code
          # never loads into an older controller (a restart would then fail).
          deployed=$("$(readlink -f "/proc/$(systemctl show -p MainPID --value jobshop.service)/exe")" version \
            | sed -n 's/.*(\([0-9a-f]*\)).*/\1/p')
          [ -n "$deployed" ] && git cat-file -e "$deployed^{commit}" 2>/dev/null && target=$deployed
        ''}
        before=$(git rev-parse HEAD)
        # Only fast-forward: a hand-edited checkout is left alone and logged.
        if ! git merge --ff-only --quiet "$target"; then
          echo "<4>workspace not fast-forwardable to $target, left as is"
          exit 0
        fi
        after=$(git rev-parse HEAD)
        if [ "$before" != "$after" ] && ! git diff --quiet "$before" "$after" -- ${lib.escapeShellArgs ws.livePacks}; then
          echo "packs changed $before..$after, reloading"
          ${jobshop} reload
        fi
      '';
    };
    systemd.timers.jobshop-workspace-sync = lib.mkIf (ws.url != "" && ws.livePacks != [ ]) {
      wantedBy = [ "timers.target" ];
      timerConfig = {
        OnBootSec = "2min";
        OnUnitActiveSec = ws.syncInterval;
      };
    };

    # Self-deploy: rebuild with the jobshop input pinned to the newest master
    # commit whose CI passed, then check the controller; on failure switch
    # back to the previous generation. The host flake's lock is not edited, so
    # the next manual deploy or auto-upgrade uses the locked rev until the host
    # repo bumps it.
    systemd.services.jobshop-self-deploy = lib.mkIf sd.enable {
      description = "Deploy the newest green Jobshop commit";
      after = [ "network-online.target" ];
      wants = [ "network-online.target" ];
      # A switch restarts units; this one must not be one of them mid-run.
      restartIfChanged = false;
      path = with pkgs; [
        nix
        git
        openssh
        jq
        coreutils
        util-linux
        config.system.build.nixos-rebuild
      ];
      environment.HOME = "/root";
      serviceConfig = {
        Type = "oneshot";
        StateDirectory = "jobshop-deploy";
        TimeoutStartSec = "30min";
      };
      script = ''
        set -eu
        state=${deployState}
        # The newest master commit with a successful CI run, read as the jobshop
        # user (its gh has the App token). Empty: nothing to do.
        rev=$(${pkgs.util-linux}/bin/runuser -u ${cfg.user} -- env ${
          lib.concatStringsSep " " (lib.mapAttrsToList (k: v: "${k}=${lib.escapeShellArg v}") environment)
        } PATH=${lib.makeBinPath path}:$PATH \
          gh run list --repo ${ghRepo} --branch ${sd.branch} --workflow ci --status success \
            --limit 1 --json headSha --jq '.[0].headSha // empty')
        [ -n "$rev" ] || { echo "no green ${sd.branch} run"; exit 0; }
        # The running controller's binary, not this unit's eval-time package,
        # which is the version from before the last self-deploy.
        running() { readlink -f "/proc/$(systemctl show -p MainPID --value jobshop.service)/exe"; }
        current=$("$(running)" version | sed -n 's/.*(\([0-9a-f]*\)).*/\1/p')
        case "$rev" in "$current"*) echo "already at $current"; exit 0 ;; esac
        if [ "$(cat $state/failed 2>/dev/null)" = "$rev" ]; then
          echo "$rev failed before, waiting for a newer green commit"; exit 0
        fi

        echo "deploying ${sd.inputName} $current -> $rev"
        prev=$(readlink -f /nix/var/nix/profiles/system)
        if ! nixos-rebuild switch --flake ${lib.escapeShellArg sd.hostFlake} --refresh \
             --override-input ${sd.inputName} ${lib.escapeShellArg "${sd.inputUrl}&rev="}"$rev"; then
          echo "<3>switch to $rev failed, rolling back"
          echo "$rev" > $state/failed
          nix-env -p /nix/var/nix/profiles/system --set "$prev"
          "$prev/bin/switch-to-configuration" switch || true
          exit 1
        fi

        # Health: the new binary answers and the controller is serving.
        ok=
        for _ in $(seq 30); do
          if systemctl is-active --quiet jobshop.service \
             && runuser -u ${cfg.user} -- env JOBSHOP_ROOT=${cfg.dataDir} "$(running)" session list >/dev/null 2>&1 \
             && "$(running)" version | grep -q "(''${rev:0:7}"; then
            ok=1; break
          fi
          sleep 2
        done
        if [ -z "$ok" ]; then
          echo "<3>controller unhealthy after $rev, rolling back to $prev"
          echo "$rev" > $state/failed
          nix-env -p /nix/var/nix/profiles/system --set "$prev"
          "$prev/bin/switch-to-configuration" switch
          exit 1
        fi
        echo "$rev" > $state/deployed
        # Read by nixos-upgrade (flags below) so the daily auto-upgrade keeps
        # this rev instead of returning to the host flake's locked one.
        echo "--override-input ${sd.inputName} ${sd.inputUrl}&rev=$rev" > $state/override
        echo "deployed $rev"
      '';
    };
    systemd.timers.jobshop-self-deploy = lib.mkIf sd.enable {
      wantedBy = [ "timers.target" ];
      timerConfig = {
        OnBootSec = "10min";
        OnUnitActiveSec = sd.interval;
      };
    };
    # Expanded by the nixos-upgrade shell script; empty until the first
    # self-deploy, so auto-upgrade then uses the locked rev as before.
    system.autoUpgrade.flags = lib.mkIf sd.enable [ "$(cat ${deployState}/override 2>/dev/null)" ];
  };
}
