# Jobshop module for NixOS (system units as the jobshop user) and
# home-manager (user units of the user that imports it, JOB-22).
# Both modes share this option tree and the jobshop.yaml generation.
# Units are collected in the internal services.jobshop.units option and
# mapped to systemd.* or systemd.user.* at the end.
# Host-specific values (limits, secrets, agent tooling) are set by the importing host.
self: mode:
{
  config,
  options,
  lib,
  pkgs,
  ...
}:
let
  isNixos = mode == "nixos";
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
        url
        ;
      token_file = g.tokenFile;
    }
    // lib.optionalAttrs (g.escalations != [ ]) { inherit (g) escalations; }
    // lib.optionalAttrs (g.poll != "") { inherit (g) poll; }
    // lib.optionalAttrs (g.deskUrl != "") { desk_url = g.deskUrl; }
  ) cfg.gateway;
  withModuleKeys =
    base:
    base
    // lib.optionalAttrs (cfg.identities != { }) {
      identities = (base.identities or { }) // cfg.identities;
    }
    // lib.optionalAttrs (livePacks != [ ]) {
      packs = (base.packs or [ ]) ++ livePacks;
    }
    // lib.optionalAttrs (cfg.gateway != { }) {
      gateway = (base.gateway or { }) // gatewaySettings;
    }
    // lib.optionalAttrs cfg.reloadOnChange {
      # The controller watches its config and packs (#533).
      reload = (base.reload or { }) // {
        watch = true;
      };
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
  # services.jobshop.store (#632): only set when the driver is chosen.
  storeAttrs = lib.optionalAttrs (cfg.store.driver != null) {
    store = {
      inherit (cfg.store) driver;
    }
    // lib.optionalAttrs (cfg.store.dsnFile != null) { dsn_file = toString cfg.store.dsnFile; };
  };
  # Journal compression (#631): only set when after is.
  jc = cfg.journal.compress;
  compressAttrs = lib.optionalAttrs (jc.after != "") {
    journal.compress = {
      inherit (jc) after level;
    };
  };
  # limits.cgroup (JOB-91): sessions get a CPU cap below the slice's and a
  # lower weight than the controller, so session load cannot starve
  # admission, ingest or dispatch. Needs the delegated cgroup, so only with
  # isolation.enable. The default cap is the slice quota minus one CPU.
  quotaPct =
    if lib.hasSuffix "%" cfg.cpuQuota then
      lib.toIntBase10 (lib.removeSuffix "%" cfg.cpuQuota)
    else
      null;
  sessionsCpu =
    if cfg.sessionsCpu != null then
      cfg.sessionsCpu
    else if quotaPct != null then
      # Whole CPUs as an integer string; a fractional quota keeps its decimals.
      let
        m = lib.max 100 (quotaPct - 100);
      in
      if lib.mod m 100 == 0 then toString (m / 100) else toString (m / 100.0)
    else
      "";
  cgroupAttrs = lib.optionalAttrs cfg.isolation.enable {
    limits.cgroup =
      lib.optionalAttrs (sessionsCpu != "") { sessions_cpu = sessionsCpu; }
      // lib.optionalAttrs (cfg.sessionsCpuWeight != null) {
        sessions_cpu_weight = cfg.sessionsCpuWeight;
      }
      // lib.optionalAttrs (cfg.controllerCpuWeight != null) {
        controller_cpu_weight = cfg.controllerCpuWeight;
      };
  };
  moduleAttrs = lib.recursiveUpdate compressAttrs (
    telemetryAttrs
    // storeAttrs
    // cgroupAttrs
    // lib.optionalAttrs (vlUrl != "") {
      journal.victorialogs = {
        url = vlUrl;
        retention_hint = vl.retentionPeriod;
      };
    }
  );
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
  systemctl = "${pkgs.systemd}/bin/systemctl" + lib.optionalString (!isNixos) " --user";
  # Root units run jobshop commands as the jobshop user; user units
  # already are that user (runuser refuses non-root callers).
  asUser = lib.optionalString isNixos "${pkgs.util-linux}/bin/runuser -u ${cfg.user} -- ";
  # tmpfiles owner: user tmpfiles create as the user already.
  own = if isNixos then "${cfg.user} ${cfg.group}" else "- -";

  # What the controller and the module's own units exec (JOB-23): git for
  # worktrees and workspace sync, sh for hooks and steps, diff in the sync
  # script, ssh for a workspace's first clone over git@. systemd adds
  # coreutils, findutils, grep and sed to every unit's PATH.
  baseTools = with pkgs; [
    git
    bash
    coreutils
    diffutils
    openssh
    # What any agent shell assumes is there (#780): without them a tool
    # call like `$(dirname $(which go))` fails and the agent blames its
    # env. NixOS has them in /run/current-system/sw/bin, which is not on
    # a unit's PATH.
    gawk
    which
    procps
    util-linux
    gnutar
    gzip
  ];
  # The old consumer toolset, kept behind defaultPath (default false).
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
    ++ baseTools
    ++ lib.optional cfg.sandbox.enable pkgs.bubblewrap
    # newuidmap/newgidmap must be the setuid wrappers, not the store
    # binaries; /run/wrappers/bin is on the service PATH below.
    ++ lib.optional cfg.isolation.enable pkgs.passt
    # pasta's deny list runs nft (JOB-19).
    ++ lib.optional cfg.isolation.enable pkgs.nftables;

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
    # Agent toolchain env (mise, nix-ld) is the consumer's, in
    # environment (JOB-23; example in docs/nixos.md).
    // cfg.environment
    // {
      JOBSHOP_ROOT = cfg.dataDir;
      HOME = cfg.dataDir;
    };

  ghRepoOf =
    url:
    lib.removeSuffix ".git" (
      lib.last (
        lib.splitString "github.com/" (lib.replaceStrings [ "github.com:" ] [ "github.com/" ] url)
      )
    );
  # Pools in settings with isolation.user.per_session (#540); each one's
  # [map_uid, map_uid+range) and gids must sit inside the jobshop user's
  # subuid/subgid range, else the pool is quarantined at its first spawn.
  # The user comes from the pool's own isolation or the last of its
  # isolation.profiles that sets one (JOB-83). Pools that get theirs from a
  # label rule or a pack are not checked here: the controller quarantines
  # them at the first spawn instead.
  isoUser =
    p:
    let
      iso = cfg.settings.pools.${p}.isolation or { };
      profs = cfg.settings.isolation_profiles or { };
      fromProfiles = lib.filter (u: u != null) (
        map (n: (profs.${n} or { }).user or null) (iso.profiles or [ ])
      );
    in
    iso.user or (if fromProfiles == [ ] then null else lib.last fromProfiles);
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
      ${lib.optionalString (
        cfg.reportProject != ""
      ) ''set -- "$@" --project ${lib.escapeShellArg cfg.reportProject}''}
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
        # Removed with self-deploy (JOB-74); kept so a host that still sets it
        # gets this message instead of an unknown-option error.
        selfDeploy = lib.mkOption {
          type = lib.types.bool;
          default = false;
          visible = false;
        };
        followInstalled = lib.mkOption {
          type = lib.types.bool;
          default = false;
          description = ''
            Hold the checkout at the installed jobshop binary's commit instead of
            origin, so a live pack that needs newer code never loads into an older
            controller (#227). For a host that deploys jobshop from this repo; the
            deploy itself is the host's (JOB-74).
          '';
        };
      };
    };

  # One service in NixOS form (the subset this module uses), so the same
  # definitions become system units or home-manager user units.
  serviceUnit = {
    options = {
      description = lib.mkOption {
        type = lib.types.str;
        default = "";
      };
      wantedBy = lib.mkOption {
        type = lib.types.listOf lib.types.str;
        default = [ ];
      };
      before = lib.mkOption {
        type = lib.types.listOf lib.types.str;
        default = [ ];
      };
      after = lib.mkOption {
        type = lib.types.listOf lib.types.str;
        default = [ ];
      };
      wants = lib.mkOption {
        type = lib.types.listOf lib.types.str;
        default = [ ];
      };
      requires = lib.mkOption {
        type = lib.types.listOf lib.types.str;
        default = [ ];
      };
      path = lib.mkOption {
        type = lib.types.listOf (lib.types.either lib.types.package lib.types.str);
        default = [ ];
      };
      environment = lib.mkOption {
        type = lib.types.attrsOf lib.types.str;
        default = { };
      };
      serviceConfig = lib.mkOption {
        type = lib.types.attrsOf lib.types.anything;
        default = { };
      };
      script = lib.mkOption {
        type = lib.types.lines;
        default = "";
      };
      reloadTriggers = lib.mkOption {
        type = lib.types.listOf lib.types.anything;
        default = [ ];
      };
      restartTriggers = lib.mkOption {
        type = lib.types.listOf lib.types.anything;
        default = [ ];
      };
      restartIfChanged = lib.mkOption {
        type = lib.types.bool;
        default = true;
      };
      stopIfChanged = lib.mkOption {
        type = lib.types.bool;
        default = true;
      };
    };
  };

  # home-manager rendering of the units. A user manager has no
  # network-online.target and starts default.target, not multi-user.target.
  hmTarget = t: if t == "multi-user.target" then "default.target" else t;
  hmDeps = l: map hmTarget (lib.filter (t: t != "network-online.target") l);
  # Environment= takes one quoted K=V per line; % is a unit specifier.
  hmQuote = s: ''"${lib.replaceStrings [ "\\" ''"'' "%" ] [ "\\\\" ''\"'' "%%" ] s}"'';
  hmService = name: s: {
    Unit =
      lib.filterAttrs (_: v: v != [ ] && v != "") {
        Description = s.description;
        Before = hmDeps s.before;
        After = hmDeps s.after;
        Wants = hmDeps s.wants;
        Requires = hmDeps s.requires;
        X-Reload-Triggers = map toString s.reloadTriggers;
        X-Restart-Triggers = map toString s.restartTriggers;
      }
      // lib.optionalAttrs (!s.restartIfChanged) { X-SwitchMethod = "keep-old"; };
    Service =
      removeAttrs s.serviceConfig [
        "User"
        "Group"
      ]
      // {
        # The same PATH NixOS gives a system unit.
        Environment = lib.mapAttrsToList (k: v: hmQuote "${k}=${v}") (
          s.environment
          // {
            PATH = "${
              lib.makeBinPath (
                s.path
                ++ (with pkgs; [
                  coreutils
                  findutils
                  gnugrep
                  gnused
                  systemd
                ])
              )
            }:${lib.makeSearchPathOutput "bin" "sbin" s.path}";
          }
        );
      }
      // lib.optionalAttrs (s.script != "") {
        ExecStart = "${pkgs.writeShellScript "${name}-start" "set -e\n${s.script}"}";
      };
    Install.WantedBy = hmDeps s.wantedBy;
  };
  hmSimple =
    kind: name: u:
    {
      Unit.Description = u.description or "";
      ${kind} = u."${lib.toLower kind}Config" or { };
    }
    // lib.optionalAttrs (u ? wantedBy) { Install.WantedBy = hmDeps u.wantedBy; };

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
          ${lib.optionalString w.followInstalled ''
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
            # Compare with the last good commit, not HEAD before the merge:
            # an operator's `git pull` moves the checkout outside this unit,
            # and that commit still needs its check and a reload (JOB-90).
            before=$(cat "$last_good")
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
            if ! ${systemctl} is-active --quiet jobshop.service; then
              echo "controller not running, reload skipped"
              exit 0
            fi
            if ! ${jobshop} reload; then
              if ! ${systemctl} is-active --quiet jobshop.service; then
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
  # Self-deploy moved to the host config (JOB-74): a host deploys jobshop with
  # its own unit and sets workspaces.<name>.followInstalled.
  imports =
    # The module no longer writes sandbox config (2026-11-06): ro, rw,
    # env_keep and shared_tools go in settings, so sandbox.enable only puts
    # bubblewrap on PATH.
    map
      (
        o:
        lib.mkRemovedOptionModule [ "services" "jobshop" "sandbox" o ]
          "The jobshop module no longer writes sandbox defaults: set them in services.jobshop.settings (defaults.sandbox.ro/rw/env_keep, shared_tools or a sandbox_profiles entry). sandbox.enable only puts bubblewrap on PATH."
      )
      [
        "systemRO"
        "envKeep"
        "rw"
        "sharedTools"
        "mise"
      ]
    ++
      map
        (
          o:
          lib.mkRemovedOptionModule [ "services" "jobshop" "selfDeploy" o ]
            "Self-deploy is no longer part of the jobshop module (JOB-74): deploy jobshop from the host config and set services.jobshop.workspaces.<name>.followInstalled = true."
        )
        [
          "enable"
          "hostFlake"
          "inputName"
          "inputUrl"
          "branch"
          "interval"
          "startCommand"
          "stateDir"
        ];

  options.services.jobshop = {
    enable = lib.mkEnableOption "the Jobshop controller";

    package = lib.mkOption {
      type = lib.types.package;
      default = self.packages.${pkgs.stdenv.hostPlatform.system}.jobshop;
      description = "Jobshop package. Pinning this store path is the deployed version.";
    };

    user = lib.mkOption {
      type = lib.types.str;
      default = if isNixos then "jobshop" else config.home.username;
      defaultText = lib.literalExpression ''"jobshop" (NixOS), config.home.username (home-manager)'';
    };

    group = lib.mkOption {
      type = lib.types.str;
      default = "jobshop";
      description = "Group of the jobshop user (NixOS only).";
    };

    # Units in NixOS form; mapped to systemd.* or systemd.user.* below.
    units = lib.mkOption {
      internal = true;
      default = { };
      type = lib.types.submodule {
        options = {
          services = lib.mkOption {
            default = { };
            type = lib.types.attrsOf (lib.types.submodule serviceUnit);
          };
          timers = lib.mkOption {
            default = { };
            type = lib.types.attrsOf (lib.types.attrsOf lib.types.anything);
          };
          paths = lib.mkOption {
            default = { };
            type = lib.types.attrsOf (lib.types.attrsOf lib.types.anything);
          };
          slices = lib.mkOption {
            default = { };
            type = lib.types.attrsOf (lib.types.attrsOf lib.types.anything);
          };
          tmpfiles = lib.mkOption {
            default = [ ];
            type = lib.types.listOf lib.types.str;
          };
        };
      };
    };

    dataDir = lib.mkOption {
      type = lib.types.path;
      default = if isNixos then "/var/lib/jobshop" else "${config.xdg.stateHome}/jobshop";
      defaultText = lib.literalExpression ''"/var/lib/jobshop" (NixOS), "''${config.xdg.stateHome}/jobshop" (home-manager)'';
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

    store = {
      driver = lib.mkOption {
        type = lib.types.nullOr (
          lib.types.enum [
            "sqlite"
            "postgres"
          ]
        );
        default = null;
        description = "Ticket store backend (config store.driver, #493). Null: not set, SQLite at dataDir/jobshop.db. Copy the data first with `jobshop admin store migrate` (docs/verify/postgres-cutover.md).";
      };
      dsnFile = lib.mkOption {
        type = lib.types.nullOr lib.types.str;
        default = null;
        example = "/run/secrets/jobshop-pg-dsn";
        description = "File holding the Postgres DSN (config store.dsn_file). Kept out of the Nix store; required with driver postgres.";
      };
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

    journal.compress = {
      after = lib.mkOption {
        type = lib.types.str;
        default = "";
        example = "24h";
        description = "Compress a closed session's journal with zstd this long after it closed (journal.compress.after, #631). Empty: off.";
      };
      level = lib.mkOption {
        type = lib.types.ints.between 1 22;
        default = 3;
        description = "zstd level for journal compression (journal.compress.level).";
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

    sessionsCpu = lib.mkOption {
      type = lib.types.nullOr lib.types.str;
      default = null;
      example = "4";
      description = ''
        CPUs all session hosts and agents may use together (limits.cgroup.sessions_cpu,
        JOB-91). Needs isolation.enable. The controller keeps the rest of cpuQuota for
        admission, ingest and dispatch. Null: cpuQuota minus one CPU, at least one;
        a cpuQuota that is not a percentage then sets no cap.
      '';
    };

    sessionsCpuWeight = lib.mkOption {
      type = lib.types.nullOr (lib.types.ints.between 1 10000);
      default = null;
      description = "cpu.weight of the sessions cgroup (default 100). Needs isolation.enable.";
    };

    controllerCpuWeight = lib.mkOption {
      type = lib.types.nullOr (lib.types.ints.between 1 10000);
      default = null;
      description = "cpu.weight of the controller's cgroup, where hooks and admission run (default 1000). Needs isolation.enable.";
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
      default = isNixos;
      defaultText = lib.literalExpression "true (NixOS), false (home-manager)";
      description = "Deny private, CGNAT/tailnet, link-local and ULA ranges, like gasserver's agent slice. NixOS only: a user manager cannot attach the IP filter, so home-manager asserts it is off.";
    };

    isolation = {
      enable = lib.mkEnableOption ''
        per-pool isolation (#388): Delegate=cpu memory pids on jobshop.service,
        a subuid/subgid range for the jobshop user, and newuidmap (shadow),
        pasta (passt) and nft (nftables) on PATH. Pools opt in with isolation: and sandbox.network: pasta.
        A pool's isolation.user.per_session.range in settings gives each session its own
        mapped uid, home and credentials, freed when the session ends (#540, JOB-28)
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

    sandbox.enable = lib.mkEnableOption "bubblewrap on PATH for sandboxed pools (#262). It sets no sandbox config: pools opt in with sandbox: in settings, which also holds every ro, rw and env_keep entry";

    defaultPath = lib.mkOption {
      type = lib.types.bool;
      default = false;
      description = "Deprecated (JOB-23): put the old agent toolset (go, nix, gnumake, gcc, ripgrep, fd, jq, curl, mise, ...) on PATH. List your agents' tools in `path` instead; this option goes away in a later release.";
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
            followInstalled = true;
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
              # Matrix and Discord were removed (2026-11-06); Desk is the operator UI.
              type = lib.types.enum [ "a2a" ];
              default = "a2a";
            };
            url = lib.mkOption {
              type = lib.types.str;
              example = "https://hermes.example/a2a";
              description = "A2A peer base URL; its agent card is under it (#547).";
            };
            escalations = lib.mkOption {
              type = lib.types.listOf lib.types.str;
              default = [ ];
              example = [ "all" ];
              description = "Escalation causes sent to the A2A peer: `all`, or run, pool, wait, guard, label. Empty: none.";
            };
            poll = lib.mkOption {
              type = lib.types.str;
              default = "";
              description = "GetTask interval for a peer task. Empty: 30s.";
            };
            tokenFile = lib.mkOption {
              type = lib.types.either lib.types.str (lib.types.attrsOf lib.types.str);
              description = "File with the peer's bearer token, readable by the jobshop user (not a Nix store path), or a secret mapping such as `{ type = \"command\"; command = \"pass show a2a/peer\"; }` (#435).";
            };
            pool = lib.mkOption {
              type = lib.types.str;
              description = "A `type: chat` pool.";
            };
            deskUrl = lib.mkOption {
              type = lib.types.str;
              default = "";
              example = "https://desk.example";
              description = "Desk base URL for links in messages. Empty: no link.";
            };
          };
        }
      );
      default = { };
      description = "A2A gateways: each is written to settings.gateway.<name> and runs as jobshop-gateway-<name>.service.";
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
      favicon = lib.mkOption {
        type = lib.types.nullOr lib.types.path;
        default = null;
        description = "Image (SVG, PNG or ICO) for the browser tab icon (--favicon). Null: the built-in icon.";
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
    # with the old dir default, so existing hosts evaluate
    # unchanged and keep their checkout.
    workspace = lib.mkOption {
      type = lib.types.submodule workspaceModule;
      visible = false;
      description = "Alias of {option}`services.jobshop.workspaces.default`.";
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
            }
          ]
        );

        assertions = [
          {
            assertion = cfg.store.driver != "postgres" || cfg.store.dsnFile != null;
            message = "services.jobshop.store: driver postgres needs dsnFile.";
          }
          {
            assertion = !(useIdentities && cfg.gitIdentity.credentialHelper != "");
            message = "services.jobshop.gitIdentity: set jobshopIdentities or credentialHelper, not both.";
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
        ++ lib.mapAttrsToList (n: w: {
          assertion = !w.selfDeploy;
          message = "services.jobshop.workspaces.${n}.selfDeploy is removed (JOB-74): deploy jobshop from the host config and set followInstalled = true.";
        }) cfg.workspaces
        ++ map (p: {
          assertion = isoRangeFits p;
          message = "services.jobshop.isolation: pool ${p} per_session ids run past subIdStart + subIdCount (${toString cfg.isolation.subIdStart} + ${toString cfg.isolation.subIdCount}); raise subIdCount or lower map_uid/map_gid or range (#540).";
        }) isoPools;

        services.jobshop.units.tmpfiles = [
          "d ${cfg.dataDir} 0700 ${own} -"
          # Local agent memory (#592): the packs' default memory_dir.
          "d ${cfg.dataDir}/memory 0750 ${own} -"
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
        # Per-host GC roots keep a running session host's binary alive across
        # Nix switches. Under home-manager the host creates this dir (root owns
        # per-user/), see docs/home-manager.md.
        ++ lib.optional isNixos "d /nix/var/nix/gcroots/per-user/${cfg.user} 0755 ${own} -";

        warnings = lib.optional cfg.defaultPath "services.jobshop.defaultPath is deprecated (JOB-23): list your agents' tools in services.jobshop.path.";

        services.jobshop.units.slices.jobshop = {
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

        services.jobshop.units.services.jobshop = {
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
            # systemd starts the controller in <own>/controller itself, so a
            # restart with live sessions under <own>/sessions never needs to
            # place a process in the unit cgroup, which has subtree_control
            # set (no internal processes rule: EBUSY, JOB-21).
            DelegateSubgroup = "controller";
          };
        }
        // lib.optionalAttrs cfg.isolation.enable {
          # newuidmap/newgidmap setuid wrappers (security.wrappers, or the
          # host's /run/wrappers under home-manager).
          path = path ++ [ "/run/wrappers" ];
        };

        # Full stop at system shutdown: end every session and reap hosts. Ordered
        # after jobshop.service so it stops first, while the controller is still up.
        # Any other stop (a switch that stopped jobshop.slice, a manual stop)
        # leaves sessions running: a full stop is `jobshop shutdown` (JOB-96).
        services.jobshop.units.services.jobshop-shutdown = {
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
            ExecStop = "${pkgs.writeShellScript "jobshop-shutdown-stop" ''
              # A switch or manual stop runs while the manager is up; any
              # other state (stopping, or no answer) is treated as shutdown.
              state=$(${systemctl} is-system-running 2>/dev/null || true)
              case "$state" in
                running | degraded)
                  echo "service manager is $state, not stopping: sessions left running"
                  exit 0
                  ;;
              esac
              exec ${jobshop} shutdown
            ''}";
            TimeoutStopSec = 120;
          };
        };

      }
      (lib.optionalAttrs isNixos {
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
        systemd = {
          inherit (cfg.units)
            services
            timers
            paths
            slices
            ;
          tmpfiles.rules = cfg.units.tmpfiles;
        };
      })
      (lib.optionalAttrs (!isNixos) {
        assertions = [
          {
            assertion = !cfg.restrictEgress;
            message = "services.jobshop.restrictEgress: a user manager cannot apply IPAddressDeny; set it false and filter egress on the host or with isolation (docs/home-manager.md).";
          }
          {
            assertion = !(vl.enable && vl.url == "");
            message = "services.jobshop.journal.victorialogs.enable runs a system service; under home-manager set journal.victorialogs.url to a VictoriaLogs the host runs.";
          }
        ];
        systemd.user = {
          services = lib.mapAttrs hmService cfg.units.services;
          timers = lib.mapAttrs (hmSimple "Timer") cfg.units.timers;
          paths = lib.mapAttrs (hmSimple "Path") cfg.units.paths;
          # sd-switch stop-starts a changed active unit, and stopping a
          # slice stops every unit in it; only the slice came back (JOB-96).
          # daemon-reload applies the new limits to the running slice.
          slices = lib.mapAttrs (
            n: u: lib.recursiveUpdate (hmSimple "Slice" n u) { Unit.X-SwitchMethod = "keep-old"; }
          ) cfg.units.slices;
          tmpfiles.rules = cfg.units.tmpfiles;
        };
      })
      {
        services.jobshop.units = lib.mkMerge (lib.mapAttrsToList workspaceUnits workspaces);
      }
      (lib.mkIf cfg.desk.enable {
        assertions = [
          {
            assertion = cfg.desk.userHeader == "" || cfg.desk.trustedProxies != [ ];
            message = "services.jobshop.desk.userHeader needs desk.trustedProxies";
          }
        ];
        # An admin-socket client like the gateway.
        services.jobshop.units.services.jobshop-desk = {
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
              ++ lib.optionals (cfg.desk.favicon != null) [
                "--favicon"
                "${cfg.desk.favicon}"
              ]
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
        services.jobshop.units.paths.jobshop-config = {
          description = "Watch the user-managed jobshop.yaml";
          wantedBy = [ "multi-user.target" ];
          pathConfig = {
            PathChanged =
              # With a config repo the sync unit checks and reloads the
              # repo's file (#532); only the module file is watched here.
              lib.optional (!repoConfig) cfg.configFile ++ [ "${cfg.dataDir}/jobshop.module.yaml" ];
            Unit = "jobshop-config.service";
          };
        };
        services.jobshop.units.services.jobshop-config = {
          description = "Check and reload a changed jobshop.yaml";
          environment.JOBSHOP_ROOT = cfg.dataDir;
          serviceConfig.Type = "oneshot";
          script = ''
            if ! out=$(${asUser}${jobshop} config check --file ${rootConfig} 2>&1); then
              echo "<3>jobshop.yaml check failed, keeping the running config: $out"
              exit 1
            fi
            echo "config check: $out"
            ${systemctl} try-reload-or-restart jobshop.service
          '';
        };
      })
      (lib.mkIf (vl.enable && vl.url == "") (
        lib.mkMerge [
          (lib.optionalAttrs isNixos {
            services.victorialogs = {
              enable = true;
              inherit (vl) listenAddress;
              extraOptions = [ "-retentionPeriod=${vl.retentionPeriod}" ];
            };
          })
          {
            services.jobshop.units.services.jobshop = {
              wants = [ "victorialogs.service" ];
              after = [ "victorialogs.service" ];
            };
          }
        ]
      ))
      (lib.mkIf cfg.ingest.enable {
        # Needs no controller or database: only the spool dir and secrets.
        services.jobshop.units.services.jobshop-ingest = {
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
        services.jobshop.units.services.jobshop-ingest-work = {
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
        services.jobshop.units.services = lib.mapAttrs' (
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
