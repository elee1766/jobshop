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
      map (
        p:
        {
          path = "${w.dir}/${p}";
          vars = w.livePackVars;
        }
        // lib.optionalAttrs (w.livePackPrefix != "") { prefix = w.livePackPrefix; }
      ) w.livePacks
    ) workspaces
  );
  # services.jobshop.gateway.<name> entries go to settings.gateway.<name>.
  gatewaySettings = lib.mapAttrs (
    _: g:
    {
      inherit (g)
        network
        pool
        rooms
        senders
        ;
      token_file = g.tokenFile;
    }
    // lib.optionalAttrs (g.homeserver != "") { inherit (g) homeserver; }
    // lib.optionalAttrs (g.interruptMarker != "") { interrupt_marker = g.interruptMarker; }
    // lib.optionalAttrs (g.escalationRoom != "") { escalation_room = g.escalationRoom; }
    // lib.optionalAttrs (g.attentionDigest != "") { attention_digest = g.attentionDigest; }
    // lib.optionalAttrs (g.artifactRoom != "") { artifact_room = g.artifactRoom; }
    // lib.optionalAttrs (g.deskUrl != "") { desk_url = g.deskUrl; }
  ) cfg.gateway;
  withModuleKeys =
    base:
    base
    // lib.optionalAttrs (cfg.identities != { }) {
      identities = (base.identities or { }) // cfg.identities;
    }
    // lib.optionalAttrs (livePacks != [ ] || deployPack != [ ]) {
      packs = (base.packs or [ ]) ++ livePacks ++ deployPack;
    }
    // lib.optionalAttrs (cfg.gateway != { }) {
      gateway = (base.gateway or { }) // gatewaySettings;
    }
    // lib.optionalAttrs cfg.reloadOnChange {
      # The controller watches its config and packs (#533).
      reload = (base.reload or { }) // {
        watch = true;
      };
    }
    // lib.optionalAttrs cfg.sandbox.enable {
      defaults = (base.defaults or { }) // {
        sandbox = {
          backend = "bubblewrap";
          inherit (cfg.sandbox) rw;
        }
        // ((base.defaults or { }).sandbox or { });
      };
      shared_tools =
        (base.shared_tools or [ ])
        ++ cfg.sandbox.sharedTools
        ++ lib.optional sharedMise cfg.sandbox.mise.dir;
    };
  # Managed mode (configFile = null): jobshop.yaml is generated from
  # settings plus the module's keys. User mode (#531): the user owns
  # jobshop.yaml and includes jobshop.module.yaml, which holds only the
  # module's keys.
  # A workspace whose checkout holds jobshop.yaml (#532) is user mode too:
  # its file is the user's file, and its sync unit checks and reloads it.
  configWorkspaces = lib.attrNames (lib.filterAttrs (_: w: w.config != null) workspaces);
  repoConfig = configWorkspaces != [ ];
  repoConfigFile =
    let
      w = workspaces.${lib.head configWorkspaces};
    in
    "${w.dir}/${w.config}";
  userFile = if repoConfig then repoConfigFile else cfg.configFile;
  configRepoState = "${cfg.dataDir}/config-repo";
  userConfig = cfg.configFile != null || repoConfig;
  telemetryAttrs = lib.optionalAttrs (telemetrySettings != { }) { telemetry = telemetrySettings; };
  # Journal search (#530): the local VictoriaLogs, or an external url.
  vl = cfg.journal.victorialogs;
  vlUrl = if vl.url != "" then vl.url else lib.optionalString vl.enable "http://${vl.listenAddress}";
  moduleAttrs =
    telemetryAttrs
    // lib.optionalAttrs (vlUrl != "") {
      journal.victorialogs = {
        url = vlUrl;
        retention_hint = vl.retentionPeriod;
      };
    };
  generatedConfig = settingsFormat.generate "jobshop.yaml" (
    lib.recursiveUpdate (withModuleKeys cfg.settings) moduleAttrs
  );
  moduleFile = settingsFormat.generate "jobshop.module.yaml" (
    lib.recursiveUpdate (withModuleKeys { }) moduleAttrs
  );
  # What a switch changes: reload and restart triggers.
  configFile = if userConfig then moduleFile else generatedConfig;
  rootConfig = "${cfg.dataDir}/jobshop.yaml";
  jobshop = lib.getExe cfg.package;
  sharedMise = cfg.sandbox.enable && cfg.sandbox.mise.tools != [ ];

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
    mise
    # From this flake's nixpkgs, so agents lint with the version CI uses.
    self.packages.${pkgs.stdenv.hostPlatform.system}.golangci-lint
  ];
  useIdentities = cfg.gitIdentity.jobshopIdentities;
  # gh through `jobshop gh` (#349): it adds the identity token for the repo
  # it acts on (--repo/-R, else the cwd's origin), asks the guard hooks
  # before writes and records them in the activity feed. An explicit
  # GH_TOKEN wins. JOBSHOP_GH names the real gh so it never finds itself.
  ghWrapper = pkgs.writeShellScriptBin "gh" ''
    JOBSHOP_GH=${lib.getExe pkgs.gh} exec ${jobshop} gh "$@"
  '';
  path =
    lib.optional useIdentities ghWrapper
    ++ cfg.path
    ++ lib.optionals cfg.defaultPath defaultTools
    ++ lib.optional cfg.sandbox.enable pkgs.bubblewrap
    # newuidmap/newgidmap must be the setuid wrappers, not the store
    # binaries; /run/wrappers/bin is on the service PATH below.
    ++ lib.optional cfg.isolation.enable pkgs.passt;

  credentialHelper =
    if useIdentities then "!${jobshop} identity git-credential" else cfg.gitIdentity.credentialHelper;
  # git reads GIT_CONFIG_KEY_n/VALUE_n, so this needs no gitconfig file.
  gitConfig =
    lib.optionals (credentialHelper != "") [
      [
        "credential.https://github.com.helper"
        credentialHelper
      ]
      [
        "url.https://github.com/.insteadOf"
        "git@github.com:"
      ]
    ]
    # The repo path picks the identity, so two repos can use two identities.
    ++ lib.optional useIdentities [
      "credential.https://github.com.useHttpPath"
      "true"
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
    # mise inside the sandbox: prebuilt runtimes through nix-ld (#262).
    // lib.optionalAttrs cfg.sandbox.enable {
      MISE_ALL_COMPILE = "false";
      MISE_PYTHON_COMPILE = "false";
      NIX_LD_LIBRARY_PATH = "/run/current-system/sw/share/nix-ld/lib";
      NIX_LD = "/run/current-system/sw/share/nix-ld/lib/ld.so";
      MISE_TRUSTED_CONFIG_PATHS = "${cfg.dataDir}/sessions:${cfg.dataDir}/workspaces";
    }
    # Sessions use the host's installs before their own (#316).
    // lib.optionalAttrs sharedMise {
      MISE_SHARED_INSTALL_DIRS = "${cfg.sandbox.mise.dir}/installs";
    }
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
  # Pools in settings with isolation.user.per_session (#540); each one's
  # [map_uid, map_uid+range) and gids must sit inside the jobshop user's
  # subuid/subgid range, else the pool is quarantined at its first spawn.
  isoUser = p: ((cfg.settings.pools.${p}.isolation or { }).user or null);
  isoPools = lib.filter (p: (isoUser p) != null && (isoUser p) ? per_session) (
    lib.attrNames (cfg.settings.pools or { })
  );
  isoRangeFits =
    p:
    let
      u = isoUser p;
      n = u.per_session.range or 0;
      start = cfg.isolation.subIdStart;
      end = start + cfg.isolation.subIdCount;
      fits = id: id >= start && id + n <= end;
    in
    !cfg.isolation.enable || (fits (u.map_uid or 0) && fits (u.map_gid or 0));
  # The Deploys view and its state command (#577), from the module's
  # source so it needs no workspace.
  deployPack = lib.optional cfg.selfDeploy.enable {
    path = "${../packs/deploy}";
    vars.state_dir = deployState;
  };
  # Deploy tools report a problem as a needs-human ticket with a stable
  # key and close it when it clears (#577); nothing in the controller
  # reads deploy state. `report <key> <title> [body]` upserts,
  # `report <key> --close [note]` closes. Never fails the caller: the
  # controller may be down mid-switch, and the state files still hold
  # the outcome for the next run.
  reportFn = ''
    report() {
      key=$1; shift
      if [ "''${1:-}" = --close ]; then
        set -- ticket upsert --key "$key" --close -b "''${2:-}"
      else
        set -- ticket upsert --key "$key" -l needs-human -b "''${2:-}" "$1"
      fi
      ${lib.optionalString (cfg.reportProject != "") ''set -- "$@" --project ${lib.escapeShellArg cfg.reportProject}''}
      as=
      [ "$(id -u)" != 0 ] || as="${pkgs.util-linux}/bin/runuser -u ${cfg.user} --"
      $as env JOBSHOP_ROOT=${cfg.dataDir} ${jobshop} "$@" >/dev/null 2>&1 \
        || echo "<4>could not report $key to jobshop (controller down?)"
    }
  '';

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
        livePackPrefix = lib.mkOption {
          type = lib.types.str;
          default = "";
          example = "dr-";
          description = ''
            Pack prefix (#184) for this workspace's live packs, so a second
            packs/github workspace gets its own pool, trigger and workflow
            names next to selfdev's.
          '';
        };
        syncInterval = lib.mkOption {
          type = lib.types.str;
          default = "2min";
          description = "How often the sync unit fast-forwards the checkout to origin and reloads the controller when it moved.";
        };
        config = lib.mkOption {
          type = lib.types.nullOr lib.types.str;
          default = null;
          example = "jobshop.yaml";
          description = ''
            Run the controller from this file in the checkout (#532): dataDir/jobshop.yaml
            links to it, and its relative pack paths resolve in the checkout. The sync unit
            checks each new commit before reloading; a failing commit is reset to the last
            good one (dataDir/config-repo.last-good) and filed as a needs-human
            `config-repo:<name>` ticket (#577), which shows in `jobshop attention`.
            The file should include jobshop.module.yaml for the module's keys and secrets.
            At most one workspace; not with configFile or settings.
          '';
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
        # After the controller, so a switch that restarts both reloads a
        # running controller (#475).
        after = [
          "${unit}.service"
          "jobshop.service"
        ];
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
            # Packs follow the installed binary, so a pack that needs newer code
            # never loads into an older controller (#227). The commit comes from
            # this generation's package, not the running process, so a stopped
            # or crash-looping controller still holds the checkout back. An
            # unknown commit leaves the checkout where it is.
            installed=$(${jobshop} version | sed -n 's/.*(\([0-9a-f]*\)).*/\1/p')
            if [ -z "$installed" ] || ! git cat-file -e "$installed^{commit}" 2>/dev/null; then
              echo "<4>workspace ${name}: installed jobshop commit '$installed' not in the checkout, not moving it"
              exit 0
            fi
            target=$installed
          ''}
          before=$(git rev-parse HEAD)
          ${lib.optionalString (w.config != null) ''
            # Config repo (#532): a commit runs only after it passes
            # `jobshop config check`; a failing one is reset to the last
            # good commit and filed as a config-repo:${name} needs-human
            # ticket (#577), closed when a later commit passes.
            ${reportFn}
            last_good=${configRepoState}.last-good
            failed=${configRepoState}.failed
            [ -s "$last_good" ] || git rev-parse HEAD > "$last_good"
            # A commit that already failed stays failed until origin moves on.
            if [ -s "$failed" ] && [ "$(head -n1 "$failed")" = "$(git rev-parse --short "$target")" ]; then
              exit 0
            fi
            if ! out=$(git merge --ff-only --quiet "$target" 2>&1); then
              rev=$(git rev-parse --short "$target")
              printf '%s\n%s\n' "$rev" "checkout not fast-forwardable (local edits or diverged history): $out" > "$failed"
              echo "<3>config repo ${name} not fast-forwardable to $target: $out"
              report config-repo:${name} "config repo ${name}: $rev not applied, checkout not fast-forwardable" \
                "$(printf 'The checkout ${w.dir} has local edits or diverged history, so %s was not applied:\n\n%s\n\nReset the checkout to origin; the ticket closes when a commit passes.' "$rev" "$out")"
              exit 1
            fi
            after=$(git rev-parse HEAD)
            [ "$before" != "$after" ] || exit 0
            # Through the root's link, so include: finds jobshop.module.yaml
            # beside it while packs resolve in the checkout.
            if ! out=$(${jobshop} config check --file ${rootConfig} 2>&1); then
              good=$(cat "$last_good")
              rev=$(git rev-parse --short "$after")
              printf '%s\n%s\n' "$rev" "$out" > "$failed"
              echo "<3>config repo ${name}: $after failed its check, reset to $good: $out"
              git reset --quiet --hard "$good"
              report config-repo:${name} "config repo ${name}: $rev failed its check, running the last good config" \
                "$(printf 'Commit %s failed jobshop config check; the host stays on %s.\n\n%s\n\nPush a fix; the ticket closes when a commit passes.' "$rev" "$good" "$out")"
              exit 1
            fi
            echo "$after" > "$last_good"
            rm -f "$failed"
            report config-repo:${name} --close "$(git rev-parse --short "$after") passed its check."
            echo "config repo moved $before..$after, check passed"
          ''}
          ${lib.optionalString (w.config == null) ''
          # Only fast-forward: a hand-edited checkout is left alone and logged.
          if ! git merge --ff-only --quiet "$target"; then
            echo "<4>workspace ${name} not fast-forwardable to $target, left as is"
            exit 0
          fi
          after=$(git rev-parse HEAD)
          ''}
          # Reload on any move, not only a diff under livePacks: a live pack
          # extends packs outside that list (selfdev extends ../github), and
          # a missed reload leaves the controller on old pools (#177).
          if [ "$before" != "$after" ]; then
            echo "checkout moved $before..$after, reloading"
            # A stopped controller loads packs on start, so a skipped reload
            # loses nothing. Failing here would fail a switch that stops the
            # controller while this unit runs and roll back a good deploy (#475).
            if ! ${pkgs.systemd}/bin/systemctl is-active --quiet jobshop.service; then
              echo "controller not running, reload skipped"
              exit 0
            fi
            if ! ${jobshop} reload; then
              if ! ${pkgs.systemd}/bin/systemctl is-active --quiet jobshop.service; then
                echo "controller stopped during reload, reload skipped"
                exit 0
              fi
              exit 1
            fi
          fi
        '';
      };
      timers."${unit}-sync" = lib.mkIf (w.livePacks != [ ] || w.config != null) {
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

    reportProject = lib.mkOption {
      type = lib.types.str;
      default = "";
      example = "js";
      description = ''
        Project the module's deploy and config-repo units file their needs-human
        tickets in (`deploy:rollback`, `deploy:config-check`, `config-repo:<name>`, #577).
        Empty: the config's default_project.
      '';
    };

    settings = lib.mkOption {
      type = settingsFormat.type;
      default = { };
      description = "Contents of jobshop.yaml. Unused, and must be empty, with configFile set.";
    };

    configFile = lib.mkOption {
      type = lib.types.nullOr lib.types.str;
      default = null;
      example = "/var/lib/jobshop/jobshop.yaml";
      description = ''
        A jobshop.yaml the user owns (#531). Null: generated from settings.
        A path: dataDir/jobshop.yaml links to it, or is left alone when it is
        that path, and the module never writes it. The module's own keys
        (identities, gateway, live packs, sandbox, attention, telemetry) go to
        dataDir/jobshop.module.yaml; the file should list it under
        `include:`. Edits are checked and reloaded by jobshop-config.path,
        or by the controller itself with reloadOnChange.
      '';
    };

    reloadOnChange = lib.mkOption {
      type = lib.types.bool;
      default = false;
      description = ''
        Set reload.watch: the controller watches jobshop.yaml, its includes
        and pack dirs and reloads after a change (#533). Replaces the
        jobshop-config.path unit of configFile mode.
      '';
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

    journal.victorialogs = {
      enable = lib.mkEnableOption "a local VictoriaLogs (nixpkgs services.victorialogs) for journal search, set as journal.victorialogs.url (#530)";
      listenAddress = lib.mkOption {
        type = lib.types.str;
        default = "127.0.0.1:9428";
        description = "Listen address of the local VictoriaLogs. Use another port when the host already runs one (e.g. Gas City's).";
      };
      retentionPeriod = lib.mkOption {
        type = lib.types.str;
        default = "90d";
        description = "-retentionPeriod of the local VictoriaLogs; also written as journal.victorialogs.retention_hint.";
      };
      url = lib.mkOption {
        type = lib.types.str;
        default = "";
        example = "http://logs.internal:9428";
        description = "An external VictoriaLogs base URL. Overrides the local one; leave enable off to use only this.";
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

    isolation = {
      enable = lib.mkEnableOption ''
        per-pool isolation (#388): Delegate=cpu memory pids on jobshop.service,
        a subuid/subgid range for the jobshop user, and newuidmap (shadow) and
        pasta (passt) on PATH. Pools opt in with isolation: and sandbox.network: pasta
      '';
      subIdStart = lib.mkOption {
        type = lib.types.int;
        default = 231072;
        description = "First subordinate uid and gid of the jobshop user's range (map_uid and map_gid pick from it).";
      };
      subIdCount = lib.mkOption {
        type = lib.types.int;
        default = 65536;
        description = "Size of the jobshop user's subuid and subgid range.";
      };
    };

    sandbox = {
      enable = lib.mkEnableOption "bubblewrap sandboxes for agents (#262): bubblewrap on PATH and defaults.sandbox.backend = bubblewrap";
      sharedTools = lib.mkOption {
        type = lib.types.listOf lib.types.str;
        default = [ ];
        description = "Dirs mounted read-only into every sandboxed session (settings.shared_tools).";
      };
      rw = lib.mkOption {
        type = lib.types.listOf lib.types.str;
        default = [
          "~/go"
          "~/.cache/go-build"
          "~/.npm"
          "~/.jcode"
          "~/.local/share/mise"
          "~/.cache/mise"
        ];
        description = "Read-write dirs for every sandboxed session (defaults.sandbox.rw): build caches, agent state, mise data.";
      };
      # Host-level mise installs shared read-only with sandboxed sessions (#316).
      mise = {
        tools = lib.mkOption {
          type = lib.types.listOf lib.types.str;
          default = [ ];
          example = [
            "node@24"
            "python@3.13"
          ];
          description = "mise tools installed outside the sandbox into <dir>/installs, mounted read-only into every sandboxed session and listed in MISE_SHARED_INSTALL_DIRS. Empty: no shared install unit.";
        };
        dir = lib.mkOption {
          type = lib.types.str;
          default = "/var/lib/jobshop-mise";
          description = "mise data dir for the shared installs; added to sandbox.sharedTools.";
        };
        interval = lib.mkOption {
          type = lib.types.str;
          default = "daily";
          description = "systemd OnCalendar for re-running the shared mise install.";
        };
      };
    };

    defaultPath = lib.mkOption {
      type = lib.types.bool;
      default = true;
      description = "Put the tools Jobshop's own agents need on PATH (git, go, nix, gnumake, gcc, ripgrep, fd, jq, curl, openssh, coreutils, bash, mise), before `path`.";
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
      jobshopIdentities = lib.mkOption {
        type = lib.types.bool;
        default = false;
        description = ''
          Get GitHub credentials from the controller's identities (#200, #216):
          git's helper for https://github.com is `jobshop identity git-credential`
          with useHttpPath, and `gh` on PATH is a wrapper that sets GH_TOKEN from
          `jobshop identity token <repo>`. Replaces credentialHelper.
        '';
      };
    };

    identities = lib.mkOption {
      type = lib.types.attrsOf settingsFormat.type;
      default = { };
      example = lib.literalExpression ''
        {
          elee-agent = {
            kind = "github-app";
            app_id = 12345;
            private_key_file = "/run/secrets/elee-agent.pem";
            installations.elee1766 = 67890;
            repos = [ "elee1766/*" ];
          };
          elee-linear = {
            kind = "linear-app";
            client_id_file = "/run/secrets/linear-app-id";
            client_secret_file = "/run/secrets/linear-app-secret";
            workspace = "acme";
            teams = [ "ENG" ];
          };
        }
      '';
      description = "Written to settings.identities (#200). Key and secret paths (private_key_file, token_file, client_id_file, client_secret_file, webhook_secret_file) come from host secrets and must be readable by the jobshop user; the yaml holds no secrets. Each may instead be `{ type = \"command\"; command = \"...\"; }`, whose stdout is the secret (#435).";
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
              type = lib.types.enum [
                "matrix"
                "discord"
              ];
              default = "matrix";
            };
            homeserver = lib.mkOption {
              type = lib.types.str;
              default = "";
              example = "http://127.0.0.1:6167";
              description = "Matrix client-server URL. Unused for discord.";
            };
            tokenFile = lib.mkOption {
              type = lib.types.either lib.types.str (lib.types.attrsOf lib.types.str);
              description = "File with the bot access token, readable by the jobshop user (not a Nix store path), or a secret mapping such as `{ type = \"command\"; command = \"pass show matrix/bot\"; }` (#435).";
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
            attentionDigest = lib.mkOption {
              type = lib.types.str;
              default = "";
              example = "0 9 * * *";
              description = "Cron (UTC) for a digest of `jobshop attention` in `escalationRoom`, posted only when the list is not empty. Empty: none.";
            };
            artifactRoom = lib.mkOption {
              type = lib.types.str;
              default = "";
              description = "Room that gets one line per delivered artifact (#508). Empty: none.";
            };
            deskUrl = lib.mkOption {
              type = lib.types.str;
              default = "";
              example = "https://desk.example";
              description = "Desk base URL for links in artifact lines. Empty: no link.";
            };
          };
        }
      );
      default = { };
      description = "Chat gateways: each is written to settings.gateway.<name> and runs as jobshop-gateway-<name>.service.";
    };

    # Desk (#202) as jobshop-desk.service; userHeader (#515), basePath
    # (#516) and trustedProxies for a reverse proxy in front.
    desk = {
      enable = lib.mkEnableOption "jobshop-desk.service running `jobshop desk`";
      listen = lib.mkOption {
        type = lib.types.str;
        default = "127.0.0.1:7480";
        description = "Address Desk serves on. Non-loopback adds --allow-remote.";
      };
      hosts = lib.mkOption {
        type = lib.types.listOf lib.types.str;
        default = [ ];
        description = "Extra Host header values to accept (--host).";
      };
      userHeader = lib.mkOption {
        type = lib.types.str;
        default = "";
        example = "Tailscale-User-Login";
        description = "Header an authenticating proxy sets to the user; writes record `desk:<user>` (#515). Empty: no user.";
      };
      basePath = lib.mkOption {
        type = lib.types.str;
        default = "/";
        example = "/desk/";
        description = "URL prefix Desk is served under (--base-path). The proxy may keep or strip it.";
      };
      trustedProxies = lib.mkOption {
        type = lib.types.listOf lib.types.str;
        default = [ ];
        example = [ "127.0.0.1" ];
        description = "Peer addresses or CIDRs whose userHeader and X-Forwarded-Prefix Desk trusts (--trusted-proxy). Required with userHeader.";
      };
      extraArgs = lib.mkOption {
        type = lib.types.listOf lib.types.str;
        default = [ ];
        description = "More `jobshop desk` flags.";
      };
    };

    ingest = {
      enable = lib.mkEnableOption "jobshop-ingest.service running `jobshop ingest serve` (#334) on settings.ingest.listen";
      worker = {
        enable = lib.mkEnableOption ''
          jobshop-ingest-work.service running `jobshop ingest work` (#443). Set
          settings.ingest.workers = "external" too, or the controller refuses its claims'';
      };
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
            assertion = !(useIdentities && cfg.gitIdentity.credentialHelper != "");
            message = "services.jobshop.gitIdentity: set jobshopIdentities or credentialHelper, not both.";
          }
          {
            assertion = !sd.enable || deployWorkspaces != [ ];
            message = "services.jobshop.selfDeploy.enable needs one workspace with selfDeploy = true.";
          }
          {
            assertion = !userConfig || cfg.settings == { };
            message = "services.jobshop: set configFile or settings, not both; with configFile put those keys in your file (the module's own keys go to jobshop.module.yaml).";
          }
          {
            assertion = lib.length configWorkspaces <= 1 && !(repoConfig && cfg.configFile != null);
            message = "services.jobshop: set workspaces.<name>.config on at most one workspace, and not with configFile.";
          }
        ]
        ++ map (p: {
          assertion = isoRangeFits p;
          message = "services.jobshop.isolation: pool ${p} per_session ids run past subIdStart + subIdCount (${toString cfg.isolation.subIdStart} + ${toString cfg.isolation.subIdCount}); raise subIdCount or lower map_uid/map_gid or range (#540).";
        }) isoPools;

        users.users.${cfg.user} = {
          isSystemUser = true;
          group = cfg.group;
          home = cfg.dataDir;
          createHome = true;
        }
        // lib.optionalAttrs cfg.isolation.enable {
          subUidRanges = [
            {
              startUid = cfg.isolation.subIdStart;
              count = cfg.isolation.subIdCount;
            }
          ];
          subGidRanges = [
            {
              startGid = cfg.isolation.subIdStart;
              count = cfg.isolation.subIdCount;
            }
          ];
        };
        users.groups.${cfg.group} = { };

        systemd.tmpfiles.rules = [
          "d ${cfg.dataDir} 0700 ${cfg.user} ${cfg.group} -"
          # The root's config is the generated file; L+ replaces the link on each switch.
          # With configFile the user's file is linked (never written) and the
          # module's keys go to jobshop.module.yaml (#531).
        ]
        ++ (
          if userConfig then
            [ "L+ ${cfg.dataDir}/jobshop.module.yaml - - - - ${moduleFile}" ]
            ++ lib.optional (userFile != rootConfig) "L+ ${rootConfig} - - - - ${userFile}"
          else
            [ "L+ ${rootConfig} - - - - ${configFile}" ]
        )
        ++ [
          # Per-host GC roots keep a running session host's binary alive across Nix switches.
          "d /nix/var/nix/gcroots/per-user/${cfg.user} 0755 ${cfg.user} ${cfg.group} -"
        ]
        ++ lib.optional sharedMise "d ${cfg.sandbox.mise.dir} 0755 ${cfg.user} ${cfg.group} -";

        systemd.services.jobshop-mise = lib.mkIf sharedMise {
          description = "Install shared mise tools for Jobshop sandboxes";
          wantedBy = [ "multi-user.target" ];
          before = [ "jobshop.service" ];
          after = [ "network-online.target" ];
          wants = [ "network-online.target" ];
          inherit path;
          # Outside the sandbox and outside the sessions' own mise data dir,
          # so sessions only ever read these installs.
          environment = environment // {
            MISE_DATA_DIR = cfg.sandbox.mise.dir;
            MISE_CACHE_DIR = "${cfg.sandbox.mise.dir}/cache";
            MISE_SHARED_INSTALL_DIRS = "";
          };
          serviceConfig = {
            Type = "oneshot";
            User = cfg.user;
            Group = cfg.group;
            Slice = "jobshop.slice";
            WorkingDirectory = cfg.sandbox.mise.dir;
          };
          script = "mise install --yes ${lib.escapeShellArgs cfg.sandbox.mise.tools}";
        };
        systemd.timers.jobshop-mise = lib.mkIf sharedMise {
          wantedBy = [ "timers.target" ];
          timerConfig = {
            OnCalendar = cfg.sandbox.mise.interval;
            Persistent = true;
          };
        };

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
          # A switch that changes only jobshop.yaml reloads the controller
          # (SIGHUP); otherwise it kept the old config until a manual
          # reload (#210). A binary change still restarts it.
          reloadTriggers = [ configFile ];
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
          }
          // lib.optionalAttrs cfg.isolation.enable {
            # The controller moves into <own>/controller and puts session
            # cgroups under <own>/sessions (#388).
            Delegate = "cpu memory pids";
          };
        }
        // lib.optionalAttrs cfg.isolation.enable {
          # newuidmap/newgidmap setuid wrappers (security.wrappers).
          path = path ++ [ "/run/wrappers" ];
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
            ${reportFn}
            # A failure's report, kept so a later run re-files it if the
            # controller was down this time.
            fail_report() { printf '%s\n%s\n%s\n' "$1" "$2" "$3" > $state/failed.report; report "$@"; }
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
            # deployed is only written after a good switch, so a manual
            # rollback, a host rebuild or auto-upgrade can leave it naming
            # another commit than the unit runs. Say so once (a rollback if
            # older: live packs may need the newer code, #227), then make it
            # follow the running binary (#225). The old mtime is kept: the
            # Deploys view compares it with failed, and a fix-up is not a deploy.
            last=$(cat $state/deployed 2>/dev/null || true)
            if [ -n "$last" ] && [ -n "$current" ]; then
              case "$last" in
                "$current"*) ;;
                *)
                  echo "<4>jobshop switched outside self-deploy (a rollback if older): running $current, last deployed $last; deployed rewritten"
                  touch -r $state/deployed $state/deployed.mtime
                  echo "$current" > $state/deployed
                  touch -r $state/deployed.mtime $state/deployed
                  rm -f $state/deployed.mtime
                  ;;
              esac
            fi
            case "$rev" in "$current"*) echo "already at $current"; exit 0 ;; esac
            if [ "$(cat $state/failed 2>/dev/null)" = "$rev" ]; then
              # Re-file the report in case the controller was down when the
              # failure happened; an unchanged upsert is a no-op.
              if [ -s $state/failed.report ]; then
                { read -r key; read -r title; body=$(cat); } < $state/failed.report
                report "$key" "$title" "$body"
              fi
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
            ${lib.optionalString (deployWorkspaces != [ ]) ''
              # Config check before the switch (#261): the new build loads this
              # host's jobshop.yaml with the live packs read from the new commit's
              # source. A pack var the host no longer satisfies fails here, filed
              # as deploy:config-check (#577), instead of after a switch.
              rm -f $state/failed.reason
              src=$(nix flake metadata --json --refresh ${lib.escapeShellArg "${sd.inputUrl}&rev="}"$rev" | jq -r .path)
              newpkg=$(nix build --no-link --print-out-paths "path:$src#jobshop")
              if ! check=$(${pkgs.util-linux}/bin/runuser -u ${cfg.user} -- \
                   "$newpkg/bin/jobshop" config check --file ${cfg.dataDir}/jobshop.yaml \
                   --replace ${
                     lib.escapeShellArg cfg.workspaces.${lib.head deployWorkspaces}.dir
                   }="$src" 2>&1); then
                echo "<3>config check failed for $rev, not switching: $check"
                printf '%s\n' "$check" > $state/failed.reason
                echo "$rev" > $state/failed
                fail_report deploy:config-check "self-deploy skipped $rev: config check failed" \
                  "$(printf 'Commit %s was not switched to: the host config does not load with it.\n\n%s\n\nFix the host config or the pack; the ticket closes on the next good deploy.' "$rev" "$check")"
                exit 1
              fi
              echo "config check: $check"
            ''}
            prev=$(readlink -f /nix/var/nix/profiles/system)
            if ! nixos-rebuild switch --flake ${lib.escapeShellArg sd.hostFlake} --refresh \
                 --override-input ${sd.inputName} ${lib.escapeShellArg "${sd.inputUrl}&rev="}"$rev"; then
              echo "<3>switch to $rev failed, rolling back"
              echo "$rev" > $state/failed
              nix-env -p /nix/var/nix/profiles/system --set "$prev"
              "$prev/bin/switch-to-configuration" switch || true
              fail_report deploy:rollback "self-deploy rolled back $rev: switch failed" \
                "nixos-rebuild switch to $rev failed; the host is back on $prev. See journalctl -u jobshop-self-deploy. The ticket closes on the next good deploy."
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
              fail_report deploy:rollback "self-deploy rolled back $rev: controller unhealthy" \
                "The controller did not come up healthy on $rev''${drill:+ (rollback drill)}; the host is back on $prev. See journalctl -u jobshop-self-deploy. The ticket closes on the next good deploy."
              exit 1
            fi
            echo "$rev" > $state/deployed
            report deploy:rollback --close "Deployed $rev."
            report deploy:config-check --close "Deployed $rev."
            rm -f $state/failed.report
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
      (lib.mkIf cfg.desk.enable {
        assertions = [
          {
            assertion = cfg.desk.userHeader == "" || cfg.desk.trustedProxies != [ ];
            message = "services.jobshop.desk.userHeader needs desk.trustedProxies";
          }
        ];
        # An admin-socket client like the gateway.
        systemd.services.jobshop-desk = {
          description = "Jobshop Desk";
          wantedBy = [ "multi-user.target" ];
          after = [ "jobshop.service" ];
          environment.JOBSHOP_ROOT = cfg.dataDir;
          serviceConfig = {
            User = cfg.user;
            Group = cfg.group;
            Slice = "jobshop.slice";
            WorkingDirectory = cfg.dataDir;
            ExecStart = lib.escapeShellArgs (
              [
                "${jobshop}"
                "desk"
                "--listen"
                cfg.desk.listen
                "--base-path"
                cfg.desk.basePath
              ]
              ++ lib.optional (
                !(
                  lib.hasPrefix "127." cfg.desk.listen
                  || lib.hasPrefix "localhost:" cfg.desk.listen
                  || lib.hasPrefix "[::1]" cfg.desk.listen
                )
              ) "--allow-remote"
              ++ lib.concatMap (h: [
                "--host"
                h
              ]) cfg.desk.hosts
              ++ lib.optionals (cfg.desk.userHeader != "") [
                "--user-header"
                cfg.desk.userHeader
              ]
              ++ lib.concatMap (p: [
                "--trusted-proxy"
                p
              ]) cfg.desk.trustedProxies
              ++ cfg.desk.extraArgs
            );
            Restart = "always";
            RestartSec = 5;
          };
        };
      })
      (lib.mkIf (userConfig && !cfg.reloadOnChange) {
        # A user-owned jobshop.yaml changes outside a switch, so
        # reloadTriggers never see it (#531). Check, then reload; a failed
        # check logs and the controller keeps its running config (#227).
        systemd.paths.jobshop-config = {
          description = "Watch the user-managed jobshop.yaml";
          wantedBy = [ "multi-user.target" ];
          pathConfig = {
            PathChanged =
              # With a config repo the sync unit checks and reloads the
              # repo's file (#532); only the module file is watched here.
              lib.optional (!repoConfig) cfg.configFile
              ++ [ "${cfg.dataDir}/jobshop.module.yaml" ];
            Unit = "jobshop-config.service";
          };
        };
        systemd.services.jobshop-config = {
          description = "Check and reload a changed jobshop.yaml";
          environment.JOBSHOP_ROOT = cfg.dataDir;
          serviceConfig.Type = "oneshot";
          script = ''
            if ! out=$(${pkgs.util-linux}/bin/runuser -u ${cfg.user} -- ${jobshop} config check --file ${rootConfig} 2>&1); then
              echo "<3>jobshop.yaml check failed, keeping the running config: $out"
              exit 1
            fi
            echo "config check: $out"
            ${pkgs.systemd}/bin/systemctl try-reload-or-restart jobshop.service
          '';
        };
      })
      (lib.mkIf (vl.enable && vl.url == "") {
        services.victorialogs = {
          enable = true;
          inherit (vl) listenAddress;
          extraOptions = [ "-retentionPeriod=${vl.retentionPeriod}" ];
        };
        systemd.services.jobshop = {
          wants = [ "victorialogs.service" ];
          after = [ "victorialogs.service" ];
        };
      })
      (lib.mkIf cfg.ingest.enable {
        # Needs no controller or database: only the spool dir and secrets.
        systemd.services.jobshop-ingest = {
          description = "Jobshop webhook receiver";
          wantedBy = [ "multi-user.target" ];
          after = [ "network-online.target" ];
          wants = [ "network-online.target" ];
          environment.JOBSHOP_ROOT = cfg.dataDir;
          restartTriggers = [ configFile ];
          serviceConfig = {
            User = cfg.user;
            Group = cfg.group;
            Slice = "jobshop.slice";
            WorkingDirectory = cfg.dataDir;
            ExecStart = "${jobshop} ingest serve";
            Restart = "always";
            RestartSec = 5;
          };
        };
      })
      (lib.mkIf cfg.ingest.worker.enable {
        # Talks to the controller's admin socket; runs admission hooks.
        systemd.services.jobshop-ingest-work = {
          description = "Jobshop ingest admission worker";
          wantedBy = [ "multi-user.target" ];
          after = [ "jobshop.service" ];
          environment.JOBSHOP_ROOT = cfg.dataDir;
          restartTriggers = [ configFile ];
          serviceConfig = {
            User = cfg.user;
            Group = cfg.group;
            Slice = "jobshop.slice";
            WorkingDirectory = cfg.dataDir;
            ExecStart = "${jobshop} ingest work --worker %H";
            Restart = "always";
            RestartSec = 5;
          };
        };
      })
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
