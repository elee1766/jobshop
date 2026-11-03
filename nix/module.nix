# NixOS module for running the Jobshop controller as its own user and slice.
# Host-specific values (limits, secrets, agent tooling) are set by the importing host.
self:
{
  config,
  options,
  lib,
  pkgs,
  ...
}:
let
  cfg = config.services.jobshop;
  opts = options.services.jobshop;
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
  # Workspaces with a url; the others are declared but cloned by nobody.
  workspaces = lib.filterAttrs (_: w: w.url != "") cfg.workspaces;
  livePacks = lib.concatLists (
    lib.mapAttrsToList (
      _: w:
      map (p: {
        path = "${w.dir}/${p}";
        vars = w.livePackVars;
      }) w.livePacks
    ) workspaces
  );
  # services.jobshop.gateway.<name> entries go to settings.gateway.<name>.
  gatewaySettings = lib.mapAttrs (
    _: g:
    {
      inherit (g)
        network
        homeserver
        pool
        rooms
        senders
        ;
      token_file = g.tokenFile;
    }
    // lib.optionalAttrs (g.interruptMarker != "") { interrupt_marker = g.interruptMarker; }
    // lib.optionalAttrs (g.escalationRoom != "") { escalation_room = g.escalationRoom; }
  ) cfg.gateway;
  settings =
    cfg.settings
    // lib.optionalAttrs (livePacks != [ ]) { packs = (cfg.settings.packs or [ ]) ++ livePacks; }
    // lib.optionalAttrs (cfg.gateway != { }) {
      gateway = (cfg.settings.gateway or { }) // gatewaySettings;
    };
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

  sd = cfg.selfDeploy;
  # GitHub owner/repo from a workspace URL.
  ghRepoOf =
    url:
    lib.removeSuffix ".git" (
      lib.last (
        lib.splitString "github.com/" (lib.replaceStrings [ "github.com:" ] [ "github.com/" ] url)
      )
    );
  deployWorkspaces = lib.attrNames (lib.filterAttrs (_: w: w.selfDeploy) cfg.workspaces);
  # The workspace whose default branch self-deploy follows, for `gh` in the deploy pick.
  ghRepo = lib.optionalString (deployWorkspaces != [ ]) (
    ghRepoOf cfg.workspaces.${lib.head deployWorkspaces}.url
  );
  deployState = "/var/lib/jobshop-deploy";

  workspaceModule =
    { name, config, ... }:
    {
      options = {
        url = lib.mkOption {
          type = lib.types.str;
          default = "";
          example = "https://github.com/elee1766/jobshop.git";
          description = "Repo cloned once into dir before the controller starts; pools use it as `repo`. Empty: no clone.";
        };
        dir = lib.mkOption {
          type = lib.types.str;
          default = "${cfg.dataDir}/workspaces/${name}";
          defaultText = lib.literalExpression ''"''${dataDir}/workspaces/<name>"'';
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
            Packs read from the live checkout (dir) instead of the Nix store, so merged
            prompt and pool changes apply after the workspace's sync unit without a
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
          description = "How often the sync unit fast-forwards the checkout to origin and reloads the controller when it moved.";
        };
        selfDeploy = lib.mkOption {
          type = lib.types.bool;
          default = false;
          description = ''
            This is the Jobshop repo: selfDeploy follows its default branch, and its
            live packs follow the deployed binary's commit. At most one workspace.
          '';
        };
      };
    };

  # Per-workspace clone and sync units, named jobshop-workspace-<name>.
  workspaceUnits =
    name: w:
    let
      unit = "jobshop-workspace-${name}";
    in
    {
      services.${unit} = {
        # Clone once. Fetches and pushes later go over https with the
        # credential helper (the controller's env); sshCommand is for this
        # first clone only.
        description = "Clone the Jobshop workspace ${name}";
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
          if [ ! -d ${w.dir}/.git ]; then
            mkdir -p "$(dirname ${w.dir})"
            ${lib.optionalString (w.sshCommand != "") "GIT_SSH_COMMAND=${lib.escapeShellArg w.sshCommand}"} \
              git clone ${lib.escapeShellArg w.url} ${w.dir}
          fi
          ${lib.optionalString (lib.hasPrefix "git@github.com:" w.url) ''
            git -C ${w.dir} remote set-url origin https://github.com/${ghRepoOf w.url}.git
          ''}
        '';
      };

      # Keep the main checkout at origin so live packs follow merges; reload
      # the controller when a pack file changed. Session worktrees are separate.
      services."${unit}-sync" = {
        description = "Fast-forward the Jobshop workspace ${name} and reload live packs";
        after = [ "${unit}.service" ];
        requires = [ "${unit}.service" ];
        inherit path environment;
        serviceConfig = {
          Type = "oneshot";
          User = cfg.user;
          Group = cfg.group;
        };
        script = ''
          set -eu
          cd ${w.dir}
          git fetch --quiet origin
          branch=$(git symbolic-ref --short HEAD)
          target="origin/$branch"
          ${lib.optionalString (sd.enable && w.selfDeploy) ''
            # Packs follow the deployed binary, so a pack that needs newer code
            # never loads into an older controller (a restart would then fail).
            deployed=$("$(readlink -f "/proc/$(systemctl show -p MainPID --value jobshop.service)/exe")" version \
              | sed -n 's/.*(\([0-9a-f]*\)).*/\1/p')
            [ -n "$deployed" ] && git cat-file -e "$deployed^{commit}" 2>/dev/null && target=$deployed
          ''}
          before=$(git rev-parse HEAD)
          # Only fast-forward: a hand-edited checkout is left alone and logged.
          if ! git merge --ff-only --quiet "$target"; then
            echo "<4>workspace ${name} not fast-forwardable to $target, left as is"
            exit 0
          fi
          after=$(git rev-parse HEAD)
          # Reload on any move, not only a diff under livePacks: a live pack
          # extends packs outside that list (selfdev extends ../github), and
          # a missed reload leaves the controller on old pools (#177).
          if [ "$before" != "$after" ]; then
            echo "checkout moved $before..$after, reloading"
            ${jobshop} reload
          fi
        '';
      };
      timers."${unit}-sync" = lib.mkIf (w.livePacks != [ ]) {
        wantedBy = [ "timers.target" ];
        timerConfig = {
          OnBootSec = "2min";
          OnUnitActiveSec = w.syncInterval;
        };
      };
    };
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

    workspaces = lib.mkOption {
      type = lib.types.attrsOf (lib.types.submodule workspaceModule);
      default = { };
      example = lib.literalExpression ''
        {
          jobshop = {
            url = "https://github.com/elee1766/jobshop.git";
            livePacks = [ "packs/selfdev" ];
            selfDeploy = true;
          };
          other.url = "https://github.com/elee1766/other.git";
        }
      '';
      description = "Repos cloned before the controller starts, each with its own live packs and sync unit.";
    };

    gateway = lib.mkOption {
      type = lib.types.attrsOf (
        lib.types.submodule {
          options = {
            network = lib.mkOption {
              type = lib.types.enum [ "matrix" ];
              default = "matrix";
            };
            homeserver = lib.mkOption {
              type = lib.types.str;
              example = "http://127.0.0.1:6167";
            };
            tokenFile = lib.mkOption {
              type = lib.types.str;
              description = "File with the bot access token, readable by the jobshop user. Not a Nix store path.";
            };
            pool = lib.mkOption {
              type = lib.types.str;
              description = "A `type: chat` pool.";
            };
            rooms = lib.mkOption {
              type = lib.types.listOf lib.types.str;
              description = "Allowed room ids; messages elsewhere are ignored.";
            };
            senders = lib.mkOption {
              type = lib.types.listOf lib.types.str;
              description = "Allowed sender ids; messages from others are ignored.";
            };
            interruptMarker = lib.mkOption {
              type = lib.types.str;
              default = "";
              description = "Prefix that sends a message with interrupt. Empty: `!`.";
            };
            escalationRoom = lib.mkOption {
              type = lib.types.str;
              default = "";
              description = "Room (one of `rooms`) that gets one new thread per escalation ticket. Empty: none.";
            };
          };
        }
      );
      default = { };
      description = "Chat gateways: each is written to settings.gateway.<name> and runs as jobshop-gateway-<name>.service.";
    };

    # Old single-workspace option. Its definitions go to workspaces.default,
    # with the old dir default and selfDeploy, so existing hosts evaluate
    # unchanged and keep their checkout.
    workspace = lib.mkOption {
      type = lib.types.submodule workspaceModule;
      visible = false;
      description = "Alias of {option}`services.jobshop.workspaces.default`.";
    };

    selfDeploy = {
      enable = lib.mkEnableOption ''
        root unit that rebuilds this host with the jobshop flake input overridden to the
        newest green commit on the default branch of the workspace with selfDeploy set, then checks the controller
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

  config = lib.mkIf cfg.enable (
    lib.mkMerge [
      {
        services.jobshop.workspaces.default = lib.mkIf opts.workspace.isDefined (
          lib.mkMerge [
            (lib.mkAliasDefinitions opts.workspace)
            {
              dir = lib.mkDefault "${cfg.dataDir}/src/${lib.removeSuffix ".git" (baseNameOf cfg.workspace.url)}";
              selfDeploy = lib.mkDefault sd.enable;
            }
          ]
        );

        assertions = [
          {
            assertion = lib.length deployWorkspaces <= 1;
            message = "services.jobshop.workspaces: selfDeploy is set on more than one workspace (${lib.concatStringsSep ", " deployWorkspaces}).";
          }
          {
            assertion = !sd.enable || deployWorkspaces != [ ];
            message = "services.jobshop.selfDeploy.enable needs one workspace with selfDeploy = true.";
          }
        ];

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

            # Rollback drill (docs/verify/rollback-drill.md): with $state/drill
            # present, this deploy reports the new controller unhealthy. The
            # marker is removed first so it applies to one deploy only.
            drill=
            if [ -e $state/drill ]; then
              rm -f $state/drill
              drill=1
              echo "<4>rollback drill: $rev will be reported unhealthy"
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
              [ -z "$drill" ] || break
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
      }
      {
        systemd = lib.mkMerge (lib.mapAttrsToList workspaceUnits workspaces);
      }
      {
        # Admin-socket clients; they retry while the controller restarts.
        systemd.services = lib.mapAttrs' (
          name: _:
          lib.nameValuePair "jobshop-gateway-${name}" {
            description = "Jobshop chat gateway ${name}";
            wantedBy = [ "multi-user.target" ];
            after = [
              "jobshop.service"
              "network-online.target"
            ];
            wants = [ "network-online.target" ];
            environment.JOBSHOP_ROOT = cfg.dataDir;
            serviceConfig = {
              User = cfg.user;
              Group = cfg.group;
              Slice = "jobshop.slice";
              WorkingDirectory = cfg.dataDir;
              ExecStart = "${jobshop} gateway ${name}";
              Restart = "always";
              RestartSec = 5;
            };
          }
        ) cfg.gateway;
      }
    ]
  );
}
