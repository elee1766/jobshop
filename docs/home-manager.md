# home-manager

2026-11-05, JOB-22.

The flake exports `homeManagerModules.default`. It takes the same
`services.jobshop` options as the NixOS module ([nixos.md](nixos.md))
and writes the same `jobshop.yaml`, but runs every unit as a systemd
user service of the user that imports it. No unit runs as root.

```nix
# flake.nix of the host config
homeConfigurations.jobcity = home-manager.lib.homeManagerConfiguration {
  pkgs = nixpkgs.legacyPackages.x86_64-linux;
  modules = [
    jobshop.homeManagerModules.default
    {
      home.username = "jobcity";
      home.homeDirectory = "/home/jobcity";
      home.stateVersion = "26.05";
      services.jobshop = {
        enable = true;
        workspaces.jobshop = {
          url = "https://github.com/gfx-labs/jobshop.git";
          livePacks = [ "packs/selfdev" ];
        };
      };
    }
  ];
};
```

## What differs from NixOS

- **Paths.** `dataDir` defaults to `$XDG_STATE_HOME/jobshop`
  (`~/.local/state/jobshop`). Sockets stay under `dataDir`.
- **Units.** The same names (`jobshop`, `jobshop-ingest`,
  `jobshop-ingest-work`, `jobshop-desk`, `jobshop-workspace-<name>`,
  ...) under `systemctl --user`. They start with
  `default.target`. A user manager has no `network-online.target`, so
  that ordering is dropped.
- **Secrets.** Secret options take any file the user can read, such as
  `~/secrets/gh`. Nothing assumes `/run/secrets`.
- **restrictEgress.** Off by default. A user manager cannot attach
  `IPAddressDeny`, so setting it is an error. Filter egress on the
  host, or use per-pool isolation with pasta and a deny list.
- **VictoriaLogs.** `journal.victorialogs.enable` runs a system service,
  so it is refused. Point `journal.victorialogs.url` at one the host runs.

## What the host provides (root, once)

- The user, with linger so its units run without a login:
  `users.users.jobcity = { isNormalUser = true; linger = true; };`
- For isolation: a subuid/subgid range for the user
  (`subUidRanges`/`subGidRanges`, matching `isolation.subIdStart` and
  `subIdCount`), and the setuid `newuidmap`/`newgidmap` in
  `/run/wrappers` (NixOS has them by default). `bwrap`, `pasta` and
  `nft` come from the module's PATH. The home may stay 0700 with no ACL:
  sandboxed mounts are made as the jobshop user (acphost#162, JOB-53).
- cgroup delegation: systemd delegates cpu, memory and pids to
  `user@.service` by default. With isolation on, limited agents run in
  transient scopes the user manager makes in `jobshop-sessions.slice`
  (`isolation.cgroupDriver = "systemd"`, the default here, #944), so
  the controller unit has no `Delegate=` and restarts with live limited
  sessions. `cgroupfs` keeps the old `Delegate=` setup.
- `/nix/var/nix/gcroots/per-user/<user>`, owned by the user, so a
  running session host's binary survives a switch (the NixOS module
  creates it; the user cannot).
- `programs.nix-ld.enable = true` if sandboxed sessions run prebuilt
  binaries (mise). Set `NIX_LD` in `environment` yourself (JOB-23).
- Group membership for any socket the host shares (agent-credd).
- Reverse proxy routes to the user's Desk and ingest ports.

Moving gasserver from the system module: [verify/system-to-user.md](verify/system-to-user.md).
