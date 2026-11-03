# Live packs with a prefix in the NixOS module (#213)

2026-11-03. `services.jobshop.workspaces.<name>.livePackPrefix` (default
"") is copied as `prefix:` onto that workspace's entries in the
generated `packs:` list (#184), so a second packs/github workspace gets
`<prefix>dev`, `<prefix>pr-work` and so on next to selfdev's names.

Host example for the second-repo check:

    workspaces.drill = {
      url = "https://github.com/elee1766/<drill-repo>.git";
      livePacks = [ "packs/github" ];
      livePackPrefix = "dr-";
      livePackVars = { repo = "elee1766/<drill-repo>"; agent = "..."; human = "elee1766"; check_cmd = "..."; };
    };

The live delivery (repo creation, host switch, one issue merged there)
needs the operator; see #213 and docs/verify/pack-prefix.md.
