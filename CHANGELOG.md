# Changelog

## [0.2.3](https://github.com/elee1766/jobshop/compare/v0.2.2...v0.2.3) (2026-11-06)


### Bug fixes

* **control:** clear a wait and add its comment in one transaction (JOB-39) ([#750](https://github.com/elee1766/jobshop/issues/750)) ([4b8cccd](https://github.com/elee1766/jobshop/commit/4b8cccdf1575ce6cb0e781a8984b6cb522f57e41))

## [0.2.2](https://github.com/elee1766/jobshop/compare/v0.2.1...v0.2.2) (2026-11-06)


### Bug fixes

* **control:** stop connect retries on SpawnError, drop journaltail wrapper (acphost[#125](https://github.com/elee1766/jobshop/issues/125)) ([#721](https://github.com/elee1766/jobshop/issues/721)) ([2f40952](https://github.com/elee1766/jobshop/commit/2f40952fd716e418a9fa44ff660389af7068db26))
* **control:** wait for the wait-ended comment in TestWaitWebhookResume ([#745](https://github.com/elee1766/jobshop/issues/745)) ([#746](https://github.com/elee1766/jobshop/issues/746)) ([0f384ab](https://github.com/elee1766/jobshop/commit/0f384ab064e61e9e5270f55648d08df7947fc980))


### Performance

* **control:** load held tickets once in session.list ([#741](https://github.com/elee1766/jobshop/issues/741)) ([39ca616](https://github.com/elee1766/jobshop/commit/39ca616beb8543de08c15e33a1fc732c2a08bbe8))

## [0.2.1](https://github.com/elee1766/jobshop/compare/v0.2.0...v0.2.1) (2026-11-06)


### Bug fixes

* **desk:** question chats offer answer text, not numbered options ([#740](https://github.com/elee1766/jobshop/issues/740)) ([aadb0ca](https://github.com/elee1766/jobshop/commit/aadb0ca7b40185cfda3ae01205b9cce1fd635821))


### Performance

* **store:** serve sqlite reads from a separate reader pool ([#742](https://github.com/elee1766/jobshop/issues/742)) ([41084fa](https://github.com/elee1766/jobshop/commit/41084fae62cac15073602b776c6e8eb2331d1995))

## [0.2.0](https://github.com/elee1766/jobshop/compare/v0.1.9...v0.2.0) (2026-11-05)


### ⚠ BREAKING CHANGES

* **nix:** hosts that relied on the default agent tools must list them in services.jobshop.path (gasserver needs that before deploy). Tested: module eval checks (default PATH has no go/gcc/mise, the warnings fire), nix fmt.

### Features

* **nix:** drop consumer-specific module defaults (JOB-23) ([#720](https://github.com/elee1766/jobshop/issues/720)) ([f725344](https://github.com/elee1766/jobshop/commit/f7253447759aa385c68dac483ec445b76124d0b0))

## [0.1.9](https://github.com/elee1766/jobshop/compare/v0.1.8...v0.1.9) (2026-11-05)


### Features

* **control:** name the spawn error on a pool crash-loop pause (JOB-27) ([#737](https://github.com/elee1766/jobshop/issues/737)) ([e1343b4](https://github.com/elee1766/jobshop/commit/e1343b4a6bcbb7c1c4f54e8bc387a9dfe85b1967))


### Bug fixes

* **control:** delete a per-session home when the session ends (JOB-28) ([#736](https://github.com/elee1766/jobshop/issues/736)) ([75acaf4](https://github.com/elee1766/jobshop/commit/75acaf46f60ec336721a0e2e8fa84f52bc54ea46))
* **control:** skip linear issues with a run started during the stall scan (JOB-26) ([#735](https://github.com/elee1766/jobshop/issues/735)) ([1afa680](https://github.com/elee1766/jobshop/commit/1afa6804fbc19f5ab0f8e1b815b5b4a2c07bfd60))

## [0.1.8](https://github.com/elee1766/jobshop/compare/v0.1.7...v0.1.8) (2026-11-05)


### Features

* **sandbox:** pasta dns, resolv_conf and dns_check settings (JOB-20) ([#733](https://github.com/elee1766/jobshop/issues/733)) ([94fe780](https://github.com/elee1766/jobshop/commit/94fe780b9dd63eb581283d69946c2ecbe4be005f))

## [0.1.7](https://github.com/elee1766/jobshop/compare/v0.1.6...v0.1.7) (2026-11-05)


### Features

* **config:** enabled: false on pools, workflows and triggers (JOB-24) ([#731](https://github.com/elee1766/jobshop/issues/731)) ([8381017](https://github.com/elee1766/jobshop/commit/83810177e41c1c76ddcd54c51b2f8050287e977b))
* **control:** quarantine isolated pools on failed preflight at load ([#728](https://github.com/elee1766/jobshop/issues/728)) ([#730](https://github.com/elee1766/jobshop/issues/730)) ([d61d6bd](https://github.com/elee1766/jobshop/commit/d61d6bdfd8a92844822f005811f373ebf59acab3))
* **nix:** home-manager module for rootless user services (JOB-22) ([#729](https://github.com/elee1766/jobshop/issues/729)) ([74261ce](https://github.com/elee1766/jobshop/commit/74261cef11690c07358dc81d4da198afbb40a575))

## [0.1.6](https://github.com/elee1766/jobshop/compare/v0.1.5...v0.1.6) (2026-11-05)


### Bug fixes

* **control:** wait notes say when a deploy or tag condition is unmet ([#716](https://github.com/elee1766/jobshop/issues/716)) ([#718](https://github.com/elee1766/jobshop/issues/718)) ([dec84fa](https://github.com/elee1766/jobshop/commit/dec84fa71b3dccaef7f7ff73a0514b8d90cf784b))
* **nix:** start the isolated controller in its cgroup subgroup so restarts work (JOB-21) ([#717](https://github.com/elee1766/jobshop/issues/717)) ([83b7ecb](https://github.com/elee1766/jobshop/commit/83b7ecbe81b5f11e317de8b1b45bd3a229326101))

## [0.1.5](https://github.com/elee1766/jobshop/compare/v0.1.4...v0.1.5) (2026-11-05)


### Features

* **control:** mapped user home and credentials for isolated pools (JOB-19) ([#713](https://github.com/elee1766/jobshop/issues/713)) ([3fea2c5](https://github.com/elee1766/jobshop/commit/3fea2c56275ae8acf8c9c4ab4fef74470a9a6d2e))


### Bug fixes

* **control:** seed new conversation from recall index when journal is gone (JOB-14) ([#708](https://github.com/elee1766/jobshop/issues/708)) ([3525821](https://github.com/elee1766/jobshop/commit/35258219d5fe3f48d343b7a743c07d6d630c7044))


### Performance

* **control:** kill admission hook process group on timeout (JOB-16) ([#707](https://github.com/elee1766/jobshop/issues/707)) ([6caf730](https://github.com/elee1766/jobshop/commit/6caf7304e1ea867f22154eaa7a19024c1969cc3e))

## [0.1.4](https://github.com/elee1766/jobshop/compare/v0.1.3...v0.1.4) (2026-11-05)


### Features

* **cli:** report ci-failures groups failed-tests artifacts for the planner (JOB-13) ([#711](https://github.com/elee1766/jobshop/issues/711)) ([a8f77a7](https://github.com/elee1766/jobshop/commit/a8f77a7b29550282e93a37a318d5fd9bd8e2d611))

## [0.1.3](https://github.com/elee1766/jobshop/compare/v0.1.2...v0.1.3) (2026-11-05)


### Bug fixes

* **control:** stall checks per instance, decision approve on linear proposals (JOB-17) ([#709](https://github.com/elee1766/jobshop/issues/709)) ([4e5d3bf](https://github.com/elee1766/jobshop/commit/4e5d3bf496b130f43f3f232311496f7ce366e3d7))

## [0.1.2](https://github.com/elee1766/jobshop/compare/v0.1.1...v0.1.2) (2026-11-05)


### Bug fixes

* **control:** stop stall checks cascading on needs_human checks ([#704](https://github.com/elee1766/jobshop/issues/704)) ([#705](https://github.com/elee1766/jobshop/issues/705)) ([e22c0cb](https://github.com/elee1766/jobshop/commit/e22c0cb71b66dd91d2c03cd8df2b250a14b7f758))

## [0.1.1](https://github.com/elee1766/jobshop/compare/v0.1.0...v0.1.1) (2026-11-05)


### Features

* **packs/github:** route release-please prs to refinery with a trigger ([#698](https://github.com/elee1766/jobshop/issues/698)) ([#699](https://github.com/elee1766/jobshop/issues/699)) ([49734d2](https://github.com/elee1766/jobshop/commit/49734d2c0744cf907ed0fb5bd1ab7b072c0da66d))


### Bug fixes

* **release:** unpin release-as 0.1.0 ([#701](https://github.com/elee1766/jobshop/issues/701)) ([#702](https://github.com/elee1766/jobshop/issues/702)) ([e1b8231](https://github.com/elee1766/jobshop/commit/e1b8231f32856b0c65077fbb277e34ef56e31712))

## 0.1.0 (2026-11-05)


### ⚠ BREAKING CHANGES

* **store:** store.Store drops IndexRecall; RecallQuery drops Kind.
* **store:** setresources, slot and resources replace the slot methods ([#579](https://github.com/elee1766/jobshop/issues/579)) (#692)
* **store:** store.Store loses AddLabel, RemoveLabel, AddComment, Reopen (internal Go interface; no schema or CLI change).
* **store:** apply(graph) replaces create, graph and edge writes ([#579](https://github.com/elee1766/jobshop/issues/579)) (#690)
* **store:** release(outcome) and expire replace handoff, close and sweeps ([#579](https://github.com/elee1766/jobshop/issues/579)) (#689)
* **store:** thread and change pages replace edges, comments, history, lastseq ([#579](https://github.com/elee1766/jobshop/issues/579)) (#685)
* **store:** list(filter) replaces the per-shape ticket reads ([#579](https://github.com/elee1766/jobshop/issues/579)) (#682)
* **store:** inbox on a store collection ([#579](https://github.com/elee1766/jobshop/issues/579)) (#662)

### Features

* **cli:** jobshop linear label, and point prompts at it (JOB-9) ([#674](https://github.com/elee1766/jobshop/issues/674)) ([3932d7e](https://github.com/elee1766/jobshop/commit/3932d7e003ac778c8d80495e400a86e977d63d7f))
* **control:** event wait kinds and a release guard for event timers (JOB-11) ([#676](https://github.com/elee1766/jobshop/issues/676)) ([4b4b785](https://github.com/elee1766/jobshop/commit/4b4b7855a18a602c89e2e777d6650ee4072acda5))
* **ingest:** one durable queue with long-poll consumers and targeted wait wake ([#679](https://github.com/elee1766/jobshop/issues/679)) ([#687](https://github.com/elee1766/jobshop/issues/687)) ([036195b](https://github.com/elee1766/jobshop/commit/036195bc66c9e76acf13d022299cf68a8358241e))
* **journal:** compress closed session journals with zstd (JOB-1) ([#656](https://github.com/elee1766/jobshop/issues/656)) ([1324cb6](https://github.com/elee1766/jobshop/commit/1324cb609acee0a11607c93cc6022c6b5c403601))
* **packs/github:** exit for work that ends without a PR (JOB-10) ([#675](https://github.com/elee1766/jobshop/issues/675)) ([bc963cf](https://github.com/elee1766/jobshop/commit/bc963cfff97ca9088a16a18f9d6cd41334ade96b))
* **packs/github:** take Linear intake next to GitHub issues ([#652](https://github.com/elee1766/jobshop/issues/652)) ([#655](https://github.com/elee1766/jobshop/issues/655)) ([5eadcf7](https://github.com/elee1766/jobshop/commit/5eadcf7189721b7006ef9b29497bb317d3319b63))
* **packs/github:** targets as Linear parent issues with sub-issues (JOB-4) ([#672](https://github.com/elee1766/jobshop/issues/672)) ([73ad21c](https://github.com/elee1766/jobshop/commit/73ad21ca79e72184a3497c5772e3580e04a51665))


### Bug fixes

* **cli:** jobshop gh resolves the repo from GET api paths too ([#686](https://github.com/elee1766/jobshop/issues/686)) ([#688](https://github.com/elee1766/jobshop/issues/688)) ([11d43d0](https://github.com/elee1766/jobshop/commit/11d43d0bdaf3c267b51265b911beabe462c0f54a))
* **control:** command steps find the jobshop binary ([#669](https://github.com/elee1766/jobshop/issues/669)) ([#671](https://github.com/elee1766/jobshop/issues/671)) ([59bb3d4](https://github.com/elee1766/jobshop/commit/59bb3d44220ce2c947915f0aa92c8c44dcf9d9bb))
* **control:** leave deferred held tickets out of the unclosed check ([#664](https://github.com/elee1766/jobshop/issues/664)) ([#667](https://github.com/elee1766/jobshop/issues/667)) ([faf541c](https://github.com/elee1766/jobshop/commit/faf541cb28771a9bbbab144afadcd3037f285faa))
* **linear:** query one catalog kind per request, paged ([#653](https://github.com/elee1766/jobshop/issues/653)) ([#654](https://github.com/elee1766/jobshop/issues/654)) ([f046f71](https://github.com/elee1766/jobshop/commit/f046f7199c2bf3de723a984e2816144a6cc36340))
* **linear:** read team lists one connection per query ([#680](https://github.com/elee1766/jobshop/issues/680)) ([#681](https://github.com/elee1766/jobshop/issues/681)) ([cdccb4b](https://github.com/elee1766/jobshop/commit/cdccb4b4f146e5db62e3e6569f77609b36183890))
* **linear:** read teams by key so name lookups fit the complexity budget ([#683](https://github.com/elee1766/jobshop/issues/683)) ([#684](https://github.com/elee1766/jobshop/issues/684)) ([52f3a16](https://github.com/elee1766/jobshop/commit/52f3a163ef375c4c6e1b8f130da026b0b2d8162e))


### Reverts

* **packs/github:** drop curator edit 395, umbrella issues and close -r rules ([#665](https://github.com/elee1766/jobshop/issues/665)) ([#666](https://github.com/elee1766/jobshop/issues/666)) ([a716c75](https://github.com/elee1766/jobshop/commit/a716c75916bd4b01dcef11ff690066227b8b109d))


### Refactors

* **store:** apply(graph) replaces create, graph and edge writes ([#579](https://github.com/elee1766/jobshop/issues/579)) ([#690](https://github.com/elee1766/jobshop/issues/690)) ([31b18be](https://github.com/elee1766/jobshop/commit/31b18be0f7bd42014bdfffd2194f9ce2e3c691fd))
* **store:** inbox on a store collection ([#579](https://github.com/elee1766/jobshop/issues/579)) ([#662](https://github.com/elee1766/jobshop/issues/662)) ([8e77708](https://github.com/elee1766/jobshop/commit/8e77708924d7e0d36d528daf51977ce24cc8e601))
* **store:** list(filter) replaces the per-shape ticket reads ([#579](https://github.com/elee1766/jobshop/issues/579)) ([#682](https://github.com/elee1766/jobshop/issues/682)) ([754a190](https://github.com/elee1766/jobshop/commit/754a190a7bb5ce4f92f3d0f98490df4c7b24e4f3))
* **store:** release(outcome) and expire replace handoff, close and sweeps ([#579](https://github.com/elee1766/jobshop/issues/579)) ([#689](https://github.com/elee1766/jobshop/issues/689)) ([b611ebf](https://github.com/elee1766/jobshop/commit/b611ebf4c4bd23714e6257073246e89b798f64c4))
* **store:** session journals on a store collection, indexrecall gone ([#579](https://github.com/elee1766/jobshop/issues/579)) ([#693](https://github.com/elee1766/jobshop/issues/693)) ([95c091d](https://github.com/elee1766/jobshop/commit/95c091d08b66b7e99648f6a5943b46f9d1ae0432))
* **store:** setresources, slot and resources replace the slot methods ([#579](https://github.com/elee1766/jobshop/issues/579)) ([#692](https://github.com/elee1766/jobshop/issues/692)) ([e7998ce](https://github.com/elee1766/jobshop/commit/e7998ce3f1e0e4c36c3001d624e2e9b046cff46e))
* **store:** thread and change pages replace edges, comments, history, lastseq ([#579](https://github.com/elee1766/jobshop/issues/579)) ([#685](https://github.com/elee1766/jobshop/issues/685)) ([17744ea](https://github.com/elee1766/jobshop/commit/17744eafcc9b0bcc336fdeab9ca3190fe2490ef2))
* **store:** update carries labels, comment and reopen ([#579](https://github.com/elee1766/jobshop/issues/579)) ([#691](https://github.com/elee1766/jobshop/issues/691)) ([ae0f386](https://github.com/elee1766/jobshop/commit/ae0f38694fe52c9e00c37cd1fc7e647a9dbdc3a9))
