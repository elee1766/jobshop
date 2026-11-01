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
  configFile = settingsFormat.generate "jobshop.yaml" (
    lib.recursiveUpdate cfg.settings (
      lib.optionalAttrs (telemetrySettings != { }) { telemetry = telemetrySettings; }
    )
  );
  jobshop = lib.getExe cfg.package;
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
      path = cfg.path;
      environment = cfg.environment // {
        JOBSHOP_ROOT = cfg.dataDir;
        HOME = cfg.dataDir;
      };
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
  };
}
