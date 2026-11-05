# release-as is for one release only (2026-11-05, #701)

`release-as` in `release-please-config.json` forces the next release
version. It stays in effect until removed, so after v0.1.0 shipped every
master push built another 0.1.0 release PR (#700) and the `release` job
failed with "Duplicate release tag: v0.1.0".

Pin `release-as` only for the one release that needs it, then remove it
(and `bootstrap-sha`, once the manifest has a version) in the next PR.
The duplicate 0.1.0 CHANGELOG section #700 added was removed; its
entry (#699) comes back in the next release, which is computed from the
v0.1.0 tag. #700 keeps its `autorelease: tagged` label, so release-please
does not retry it.
