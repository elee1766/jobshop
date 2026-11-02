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
  ];
in
# Assertions run at evaluation, so `nix flake check --no-build` covers them;
# the pack paths in the generated file are checked when the check builds.
assert lib.all lib.id checks;
pkgs.runCommand "jobshop-module-test" { } ''
  grep -q '/var/lib/jobshop/workspaces/jobshop/packs/selfdev' ${yamlPacks two}
  grep -q '/var/lib/jobshop/workspaces/other/packs/github' ${yamlPacks two}
  grep -q 'repo: elee1766/other' ${yamlPacks two}
  touch $out
''
