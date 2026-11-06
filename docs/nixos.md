# NixOS

Moved from the README (2026-11-05, #551).

The flake exports `nixosModules.default`. It gives `services.jobshop` with
`settings` (the yaml), `workspaces.<name>` (repos cloned before start,
with live packs that follow the checkout), `gateway.<name>`,
`identities` and telemetry. Deploying jobshop itself is the host's job
(JOB-74); a host that does it sets `workspaces.<name>.followInstalled`
so live packs follow the installed binary. See
[`nix/module.nix`](../nix/module.nix), and [layout.md](layout.md) for
deployment. The sandbox option is in [isolation.md](isolation.md).
The same options as rootless user services: [home-manager.md](home-manager.md).

## Running from your own repo

Keep `jobshop.yaml` and `packs/` in a git repo and point a workspace at it:

```nix
services.jobshop.workspaces.config = {
  url = "https://github.com/me/jobshop-config.git";
  config = "jobshop.yaml";
};
```

The host links `dataDir/jobshop.yaml` to the checkout, so relative
`packs[].path` resolve in the repo. The file lists
`include: [jobshop.module.yaml]`, where the module puts identities and
secret paths: secrets stay out of the repo. Each push is fast-forwarded,
checked with `jobshop config check`, then reloaded. A commit that fails
is reset to the last good one and shows in `jobshop attention` as
`config-repo`, and so does a checkout a local edit keeps from moving.
See [notes/532-config-repo.md](notes/532-config-repo.md).

## Agent tools

The module puts on PATH only what jobshop itself runs: git, bash,
coreutils, diffutils, openssh, the `gh` wrapper with identities, bwrap
with `sandbox.enable`, pasta and nft with `isolation.enable`.
`sandbox.enable` writes no sandbox config: no backend, mounts or env
(2026-11-06). Sandboxed pools set `sandbox:` and the mounts in
`settings`. Your agents' tools, caches and env are yours to set. A host like gasserver (Go, mise, jcode,
nix-ld) writes:

```nix
programs.nix-ld.enable = true;
services.jobshop = {
  path = with pkgs; [ go_1_27 nix gnumake gcc ripgrep fd jq curl mise ];
  environment = {
    MISE_ALL_COMPILE = "false";
    MISE_PYTHON_COMPILE = "false";
    NIX_LD = "/run/current-system/sw/share/nix-ld/lib/ld.so";
    NIX_LD_LIBRARY_PATH = "/run/current-system/sw/share/nix-ld/lib";
    MISE_TRUSTED_CONFIG_PATHS = "/var/lib/jobshop/sessions:/var/lib/jobshop/workspaces";
    MISE_SHARED_INSTALL_DIRS = "/var/lib/jobshop-mise/installs";
  };
  sandbox.enable = true; # bubblewrap on PATH only
  settings = {
    # A profile pools opt into with sandbox.profiles; nothing is sandboxed by default.
    sandbox_profiles.host = {
      backend = "bubblewrap";
      ro = [ "/nix/store" "/etc" "/run/current-system" "/run/systemd/resolve" ];
      rw = [ "~/go" "~/.cache/go-build" "~/.npm" "~/.jcode" "~/.local/share/mise" "~/.cache/mise" ];
      env_keep = [ "HOME" "USER" "LOGNAME" "LANG" "LC_ALL" "TZ" "TERM" "SSL_CERT_FILE"
        "NIX_SSL_CERT_FILE" "NIX_PATH" "NIX_LD" "NIX_LD_LIBRARY_PATH" "GOPATH" "GOCACHE"
        "GOFLAGS" "GOTOOLCHAIN" "GIT_*" "MISE_*" ];
    };
    # Shared mise installs, read-only in every sandbox.
    shared_tools = [ "/var/lib/jobshop-mise" ];
    pools.dev.sandbox.profiles = [ "host" ];
  };
  };
};
systemd.tmpfiles.rules = [ "d /var/lib/jobshop-mise 0755 jobshop jobshop -" ];
systemd.services.jobshop-mise = {
  wantedBy = [ "multi-user.target" ];
  before = [ "jobshop.service" ];
  after = [ "network-online.target" ];
  wants = [ "network-online.target" ];
  path = [ pkgs.mise pkgs.coreutils ];
  environment = {
    MISE_DATA_DIR = "/var/lib/jobshop-mise";
    MISE_CACHE_DIR = "/var/lib/jobshop-mise/cache";
  };
  serviceConfig = { Type = "oneshot"; User = "jobshop"; Group = "jobshop"; };
  script = "mise install --yes node@24 python@3.13";
  startAt = "daily";
};
```

`defaultPath = true` still adds the old toolset with a warning.
`sandbox.systemRO`, `envKeep`, `rw`, `sharedTools` and `mise` are
removed: evaluation fails with a message pointing at `settings`.
