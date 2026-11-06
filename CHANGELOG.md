# Changelog

## [0.6.15](https://github.com/elee1766/jobshop/compare/v0.6.14...v0.6.15) (2026-11-06)


### Bug fixes

* **control:** linear_issue_no_run grace defaults to off ([#901](https://github.com/elee1766/jobshop/issues/901)) ([6183309](https://github.com/elee1766/jobshop/commit/61833095f24900c3a366ffc96e0350b6bc70a3de))

## [0.6.14](https://github.com/elee1766/jobshop/compare/v0.6.13...v0.6.14) (2026-11-06)


### Features

* **linear:** answer and close agent sessions; self-assign may land as delegate ([#898](https://github.com/elee1766/jobshop/issues/898)) ([008fed8](https://github.com/elee1766/jobshop/commit/008fed81006b00673b272499377b4e45ed9a8766))

## [0.6.13](https://github.com/elee1766/jobshop/compare/v0.6.12...v0.6.13) (2026-11-06)


### Features

* **control:** reserve cpu for the controller with a sessions cgroup (JOB-91) ([#894](https://github.com/elee1766/jobshop/issues/894)) ([deb0b18](https://github.com/elee1766/jobshop/commit/deb0b180d675372331da2eb64509763b96d707ca))

## [0.6.12](https://github.com/elee1766/jobshop/compare/v0.6.11...v0.6.12) (2026-11-06)


### Features

* **desk:** replaceable nav slot with placements and api.views() (JOB-86) ([#888](https://github.com/elee1766/jobshop/issues/888)) ([60ccf3d](https://github.com/elee1766/jobshop/commit/60ccf3d388348384c71a78c198477383b229ba79))

## [0.6.11](https://github.com/elee1766/jobshop/compare/v0.6.10...v0.6.11) (2026-11-06)


### Features

* **control:** configurable max_open for stall triggers (JOB-93) ([#891](https://github.com/elee1766/jobshop/issues/891)) ([266440b](https://github.com/elee1766/jobshop/commit/266440b1ecaa4b5b0f0a4bbfe7dacd0c30b5bc96))

## [0.6.10](https://github.com/elee1766/jobshop/compare/v0.6.9...v0.6.10) (2026-11-06)


### Features

* **hooks:** url guardrail hooks and configurable retry for url hooks ([#887](https://github.com/elee1766/jobshop/issues/887)) ([30a711a](https://github.com/elee1766/jobshop/commit/30a711aac26f54c8fe400c4af6267a921abb2951))

## [0.6.9](https://github.com/elee1766/jobshop/compare/v0.6.8...v0.6.9) (2026-11-06)


### Bug fixes

* **nix:** config repo sync reloads when HEAD differs from last-good (JOB-90) ([#886](https://github.com/elee1766/jobshop/issues/886)) ([d15d266](https://github.com/elee1766/jobshop/commit/d15d2663fdf7e812967c17177a97f8c5acfd215b))

## [0.6.8](https://github.com/elee1766/jobshop/compare/v0.6.7...v0.6.8) (2026-11-06)


### Features

* **desk:** --favicon replaces the tab icon ([#878](https://github.com/elee1766/jobshop/issues/878)) ([a331035](https://github.com/elee1766/jobshop/commit/a3310350451873e03e966eeab857b993a8c258ab))
* **desk:** app slot so a plugin can replace the whole frame (JOB-85) ([#881](https://github.com/elee1766/jobshop/issues/881)) ([ec1e8e2](https://github.com/elee1766/jobshop/commit/ec1e8e280f9ec567f5d6dc277cf293f744e19bb6))
* **desk:** label tool calls by intent and target; sessions report their agent ([#883](https://github.com/elee1766/jobshop/issues/883)) ([a19a480](https://github.com/elee1766/jobshop/commit/a19a480b34d37271e069b495235a22542d3739dc))
* **desk:** session console with tabs at the bottom; model in the chat footer ([#882](https://github.com/elee1766/jobshop/issues/882)) ([8b4af2c](https://github.com/elee1766/jobshop/commit/8b4af2c25f55a95bb4c0eaedbb8b63cb7017e532))


### Bug fixes

* **desk:** compact session chip; name moves to the tooltip ([#879](https://github.com/elee1766/jobshop/issues/879)) ([91890c9](https://github.com/elee1766/jobshop/commit/91890c99404e56360755f7399d472cf2fe88f300))
* **packs/github:** prefix the dev_pool default per instance ([#880](https://github.com/elee1766/jobshop/issues/880)) ([#885](https://github.com/elee1766/jobshop/issues/885)) ([3c78218](https://github.com/elee1766/jobshop/commit/3c78218d3d8e83a19425184b4c7517d0c300b09b))

## [0.6.7](https://github.com/elee1766/jobshop/compare/v0.6.6...v0.6.7) (2026-11-06)


### Features

* **desk:** plugin-defined slots to extend existing pages (JOB-87) ([#875](https://github.com/elee1766/jobshop/issues/875)) ([9eb3ac3](https://github.com/elee1766/jobshop/commit/9eb3ac3eb7b43efb68061b3a466d450c200498c1))


### Bug fixes

* **identity:** sessions renew a Linear token Linear refused ([#866](https://github.com/elee1766/jobshop/issues/866)) ([#876](https://github.com/elee1766/jobshop/issues/876)) ([d3aa5fd](https://github.com/elee1766/jobshop/commit/d3aa5fd69912f17de801e87d1b442287688b7e33))

## [0.6.6](https://github.com/elee1766/jobshop/compare/v0.6.5...v0.6.6) (2026-11-06)


### Bug fixes

* **control:** match github.checks workflow= case-insensitively and by file stem (JOB-88) ([#873](https://github.com/elee1766/jobshop/issues/873)) ([b48961e](https://github.com/elee1766/jobshop/commit/b48961ec853262dc493ccee927d78fc82f8a44a2))

## [0.6.5](https://github.com/elee1766/jobshop/compare/v0.6.4...v0.6.5) (2026-11-06)


### Bug fixes

* **linear:** renew a revoked app token on 401; me self-assigns instead of delegating ([#871](https://github.com/elee1766/jobshop/issues/871)) ([12b7c57](https://github.com/elee1766/jobshop/commit/12b7c579ae3c19159c8ac3063e4dacc106b65dce))

## [0.6.4](https://github.com/elee1766/jobshop/compare/v0.6.3...v0.6.4) (2026-11-06)


### Bug fixes

* **desk:** hide thinking in chat; a session/load replay is not new output ([#867](https://github.com/elee1766/jobshop/issues/867)) ([6f53661](https://github.com/elee1766/jobshop/commit/6f5366150d039824ab0536d23d06692316da3866))

## [0.6.3](https://github.com/elee1766/jobshop/compare/v0.6.2...v0.6.3) (2026-11-06)


### Features

* **config:** sandbox and isolation profiles by name or label (JOB-83) ([#862](https://github.com/elee1766/jobshop/issues/862)) ([24e0fe9](https://github.com/elee1766/jobshop/commit/24e0fe93ee2312c1cb6a9e6bea56dc46f0621002))

## [0.6.2](https://github.com/elee1766/jobshop/compare/v0.6.1...v0.6.2) (2026-11-06)


### Bug fixes

* **packs/github:** release-merge waits on the ci workflow, dev tickets name the dev pool ([#861](https://github.com/elee1766/jobshop/issues/861)) ([#863](https://github.com/elee1766/jobshop/issues/863)) ([5a181b2](https://github.com/elee1766/jobshop/commit/5a181b2b0ff92581f2fa29a09c695c867f09868d))

## [0.6.1](https://github.com/elee1766/jobshop/compare/v0.6.0...v0.6.1) (2026-11-06)


### Features

* **linear:** linear assign, app tokens take issues as delegate; jobshop help ([#856](https://github.com/elee1766/jobshop/issues/856)) ([1932895](https://github.com/elee1766/jobshop/commit/1932895c69077b528e45f90a89b624932ee5c60f))

## [0.6.0](https://github.com/elee1766/jobshop/compare/v0.5.2...v0.6.0) (2026-11-06)


### ⚠ BREAKING CHANGES

* **nix:** sandbox.enable only puts bubblewrap on PATH ([#853](https://github.com/elee1766/jobshop/issues/853))

### Bug fixes

* **github:** agents file Linear issues in Todo so intake picks them up ([#851](https://github.com/elee1766/jobshop/issues/851)) ([ab587e2](https://github.com/elee1766/jobshop/commit/ab587e266e42c2f9e0fa79758e7e7ef1386523f6))


### Refactors

* **nix:** sandbox.enable only puts bubblewrap on PATH ([#853](https://github.com/elee1766/jobshop/issues/853)) ([a60f495](https://github.com/elee1766/jobshop/commit/a60f4952c14b4333e24d4a41ba0f9c76468ff006))

## [0.5.2](https://github.com/elee1766/jobshop/compare/v0.5.1...v0.5.2) (2026-11-06)


### Features

* **desk:** pool labels and a pools view grouped by instance (JOB-81) ([#850](https://github.com/elee1766/jobshop/issues/850)) ([251ecff](https://github.com/elee1766/jobshop/commit/251ecff296f0144bc89381f57143e08b8b817d53))

## [0.5.1](https://github.com/elee1766/jobshop/compare/v0.5.0...v0.5.1) (2026-11-06)


### Features

* **desk:** remove the built-in artifacts view (JOB-80) ([#848](https://github.com/elee1766/jobshop/issues/848)) ([6a55767](https://github.com/elee1766/jobshop/commit/6a5576730e56bc09659cabb1ba486871d3e381c9))

## [0.5.0](https://github.com/elee1766/jobshop/compare/v0.4.4...v0.5.0) (2026-11-06)


### ⚠ BREAKING CHANGES

* **nix:** hosts that set `services.jobshop.selfDeploy` or `workspaces.<name>.selfDeploy` fail evaluation; deploy from the host config and set `followInstalled = true`.

### Refactors

* **nix:** move self-deploy out of the jobshop module (JOB-74) ([#844](https://github.com/elee1766/jobshop/issues/844)) ([2c288ef](https://github.com/elee1766/jobshop/commit/2c288efe39a7eb63cda85d5e8a144db0b2906e9e))

## [0.4.4](https://github.com/elee1766/jobshop/compare/v0.4.3...v0.4.4) (2026-11-06)


### Features

* **nix:** workspace followInstalled holds live packs at the installed binary ([#840](https://github.com/elee1766/jobshop/issues/840)) ([e3a26e8](https://github.com/elee1766/jobshop/commit/e3a26e8bc23c04fb2d097a84c5e8181e172f729c))

## [0.4.3](https://github.com/elee1766/jobshop/compare/v0.4.2...v0.4.3) (2026-11-06)


### Features

* **control:** workflow= for github.checks waits, cancelled PR checks are not red (JOB-75) ([#841](https://github.com/elee1766/jobshop/issues/841)) ([d606f9f](https://github.com/elee1766/jobshop/commit/d606f9f4790b52aa22046af98cd4ae2c5d474a11))

## [0.4.2](https://github.com/elee1766/jobshop/compare/v0.4.1...v0.4.2) (2026-11-06)


### Features

* **control:** on: and fallback: for stall triggers (JOB-72) ([#839](https://github.com/elee1766/jobshop/issues/839)) ([e7083d2](https://github.com/elee1766/jobshop/commit/e7083d2e7e288213d7c7fb519f829316fd38c775))
* **linear:** agents assign themselves when they take an issue ([#836](https://github.com/elee1766/jobshop/issues/836)) ([d05f2cb](https://github.com/elee1766/jobshop/commit/d05f2cb49942cc929777e28fbc2197debd94163e))


### Bug fixes

* **control:** checks waits on a commit ignore cancelled and skipped runs ([#831](https://github.com/elee1766/jobshop/issues/831)) ([#837](https://github.com/elee1766/jobshop/issues/837)) ([615742a](https://github.com/elee1766/jobshop/commit/615742a583ea9da892c63e01d1aaf48202df124d))

## [0.4.1](https://github.com/elee1766/jobshop/compare/v0.4.0...v0.4.1) (2026-11-06)


### Features

* **deploy:** start self-deploy on ci success event ([#826](https://github.com/elee1766/jobshop/issues/826)) ([#834](https://github.com/elee1766/jobshop/issues/834)) ([ceec294](https://github.com/elee1766/jobshop/commit/ceec2945cd565c7e506d839c6df56f2de2bce7e2))

## [0.4.0](https://github.com/elee1766/jobshop/compare/v0.3.7...v0.4.0) (2026-11-06)


### ⚠ BREAKING CHANGES

* **store:** store timestamps in nanoseconds (JOB-58) ([#821](https://github.com/elee1766/jobshop/issues/821))

### Features

* **control:** event triggers with poll fallback and a command action (JOB-62) ([#828](https://github.com/elee1766/jobshop/issues/828)) ([1374500](https://github.com/elee1766/jobshop/commit/1374500176ab28b6a50b692fa0b090d0f0317092))
* **control:** fall back to a new ACP session on unknown session id (JOB-65) ([#820](https://github.com/elee1766/jobshop/issues/820)) ([bddb7d8](https://github.com/elee1766/jobshop/commit/bddb7d855b9fe15e653d0320ad19be0ed2557a0a))
* **control:** probe the agent shell PATH before spawning a host (JOB-67) ([#819](https://github.com/elee1766/jobshop/issues/819)) ([cc0a5a2](https://github.com/elee1766/jobshop/commit/cc0a5a24b64fdf28eaff42ddd4456b984e9dff69))
* **control:** run hooks, prompt rules, desk exec, triggers and command steps through proc.Start ([#810](https://github.com/elee1766/jobshop/issues/810)) ([#823](https://github.com/elee1766/jobshop/issues/823)) ([2e662ca](https://github.com/elee1766/jobshop/commit/2e662ca3f4dd1c43ca75acf6d83a7544d96ee3c2))
* **control:** run start refuses an ineligible linear issue without --force (JOB-69) ([#816](https://github.com/elee1766/jobshop/issues/816)) ([ee42085](https://github.com/elee1766/jobshop/commit/ee420851ad4ed84403f6752477da2d91866a8cee))
* **desk:** pending state for attention item actions (JOB-60) ([#807](https://github.com/elee1766/jobshop/issues/807)) ([dc25187](https://github.com/elee1766/jobshop/commit/dc25187d6c8779becdef5fbcf72c70c3d96ef37e))
* **ids:** configurable ticket id generators (JOB-57) ([#824](https://github.com/elee1766/jobshop/issues/824)) ([fab3ff1](https://github.com/elee1766/jobshop/commit/fab3ff12145adb87c40087fc77691d8ed93ddf68))
* **packs/github:** route implement steps to dev-sonnet or dev with retry escalation (JOB-59) ([#814](https://github.com/elee1766/jobshop/issues/814)) ([fab4b68](https://github.com/elee1766/jobshop/commit/fab4b683bb3bc1ae132302968666c12e323d306d))
* **packs:** stall pool with stall_model and stall_effort vars ([#785](https://github.com/elee1766/jobshop/issues/785)) ([87f0545](https://github.com/elee1766/jobshop/commit/87f054588bb393a6142fbdd1a4ffa9b0ca54cf2f))
* **store:** store timestamps in nanoseconds (JOB-58) ([#821](https://github.com/elee1766/jobshop/issues/821)) ([a4485de](https://github.com/elee1766/jobshop/commit/a4485de6bd3140b1635ff3b8adf55a550ea14186))


### Bug fixes

* **control:** back off on provider network errors instead of crashing (JOB-66) ([#809](https://github.com/elee1766/jobshop/issues/809)) ([787ae09](https://github.com/elee1766/jobshop/commit/787ae09c63a9668ede931315afd85c77695fd566))
* **control:** keep idle sessions with background jobs, refuse release with unpushed commits (JOB-70) ([#813](https://github.com/elee1766/jobshop/issues/813)) ([bd466ad](https://github.com/elee1766/jobshop/commit/bd466ad3c9e24ada8899303bfe48260b488c51ab))
* **linear:** treat open stated dependencies as intake blockers (JOB-68) ([#815](https://github.com/elee1766/jobshop/issues/815)) ([f6f3372](https://github.com/elee1766/jobshop/commit/f6f3372f6b0dd3ebc0f5983a9f57e848fcd1987f))

## [0.3.7](https://github.com/elee1766/jobshop/compare/v0.3.6...v0.3.7) (2026-11-06)


### Features

* **desk:** link Linear issues on runs, tickets and attention rows ([#784](https://github.com/elee1766/jobshop/issues/784)) ([b3f263c](https://github.com/elee1766/jobshop/commit/b3f263c4de05a7cef97179ddc2338e1938194c2d))


### Bug fixes

* **ingest:** prune finished inbox records in batches, default 7d (JOB-63) ([#806](https://github.com/elee1766/jobshop/issues/806)) ([9c97c50](https://github.com/elee1766/jobshop/commit/9c97c50573541b4f6152f08a08c35efc3d984a74))


### Performance

* **store:** index records by state and created_at for the inbox backlog count (JOB-61) ([#804](https://github.com/elee1766/jobshop/issues/804)) ([7841352](https://github.com/elee1766/jobshop/commit/7841352eea8ab6f0efc3c90be6e7e55b51027660))

## [0.3.6](https://github.com/elee1766/jobshop/compare/v0.3.5...v0.3.6) (2026-11-06)


### Bug fixes

* **control:** check only mount sources under a sandbox, bump acphost to v0.15.0 (JOB-53) ([#801](https://github.com/elee1766/jobshop/issues/801)) ([7ecb4f6](https://github.com/elee1766/jobshop/commit/7ecb4f6f312f7252a0bba4e09e188c4176406972))

## [0.3.5](https://github.com/elee1766/jobshop/compare/v0.3.4...v0.3.5) (2026-11-06)


### Bug fixes

* **rpc:** use acphost sandbox/userns/cgroup packages instead of root aliases ([#793](https://github.com/elee1766/jobshop/issues/793)) ([#794](https://github.com/elee1766/jobshop/issues/794)) ([83e34da](https://github.com/elee1766/jobshop/commit/83e34dadbf27c186247bf75a41001b13b2c21cb0))

## [0.3.4](https://github.com/elee1766/jobshop/compare/v0.3.3...v0.3.4) (2026-11-06)


### Bug fixes

* **control:** isolation tests run on dev machines; oom at startup ends the session ([#789](https://github.com/elee1766/jobshop/issues/789)) ([31e9b74](https://github.com/elee1766/jobshop/commit/31e9b7429022e3a778a52a70dfbec9217408ecdf))
* **control:** wait for escalation ticket in TestAutoStepFailureEscalates ([#786](https://github.com/elee1766/jobshop/issues/786)) ([#787](https://github.com/elee1766/jobshop/issues/787)) ([5955be3](https://github.com/elee1766/jobshop/commit/5955be31f2f3adaa24a96030f289c82f13bc48ab))

## [0.3.3](https://github.com/elee1766/jobshop/compare/v0.3.2...v0.3.3) (2026-11-06)


### Features

* **control:** question chats answer, follow the worker and close when settled (JOB-52) ([#779](https://github.com/elee1766/jobshop/issues/779)) ([b3e772c](https://github.com/elee1766/jobshop/commit/b3e772c3826cf2d7a6cb55f9a4957965a59b31e8))
* **sandbox:** overlay mounts merging layer dirs per target (JOB-55) ([#782](https://github.com/elee1766/jobshop/issues/782)) ([0dbfbcb](https://github.com/elee1766/jobshop/commit/0dbfbcb9675c7efb1affb635501127c14e720c71))


### Bug fixes

* **nix:** put which, awk, ps and friends back on the agent shell PATH ([#780](https://github.com/elee1766/jobshop/issues/780)) ([#781](https://github.com/elee1766/jobshop/issues/781)) ([870ee15](https://github.com/elee1766/jobshop/commit/870ee15255509ebed12ee2d86614e1fdab96e16a))

## [0.3.2](https://github.com/elee1766/jobshop/compare/v0.3.1...v0.3.2) (2026-11-06)


### Features

* **control:** read-only sandbox guidance mounts (JOB-51) ([#773](https://github.com/elee1766/jobshop/issues/773)) ([3599edb](https://github.com/elee1766/jobshop/commit/3599edb8403579875677473407effc0e2886777a))
* **desk:** show the working session on in-progress cards and traces (JOB-54) ([#777](https://github.com/elee1766/jobshop/issues/777)) ([46ceb02](https://github.com/elee1766/jobshop/commit/46ceb029bd7ca36061fa1a28c82e477c144e7676))
* **packs:** refinery_model and refinery_effort vars ([#775](https://github.com/elee1766/jobshop/issues/775)) ([398e866](https://github.com/elee1766/jobshop/commit/398e8668c1e96059d8277c7bcc4a32a6632825b5))


### Bug fixes

* **nix:** self-deploy reconcile writes the full green sha ([#772](https://github.com/elee1766/jobshop/issues/772)) ([15c4670](https://github.com/elee1766/jobshop/commit/15c4670c7e7a8e63e4d27e5a670f355c9aa21ebc))

## [0.3.1](https://github.com/elee1766/jobshop/compare/v0.3.0...v0.3.1) (2026-11-06)


### Features

* **cli:** log level and format flags, context log ids, quiet tests (JOB-49) ([#769](https://github.com/elee1766/jobshop/issues/769)) ([371a8c6](https://github.com/elee1766/jobshop/commit/371a8c6546f5a2e3bd77fa9f13e693f1cac6ee95))

## [0.3.0](https://github.com/elee1766/jobshop/compare/v0.2.5...v0.3.0) (2026-11-06)


### ⚠ BREAKING CHANGES

* **gateway:** remove Matrix and Discord gateways, keep A2A and Desk ([#766](https://github.com/elee1766/jobshop/issues/766))

### Features

* **gateway:** remove Matrix and Discord gateways, keep A2A and Desk ([#766](https://github.com/elee1766/jobshop/issues/766)) ([6146655](https://github.com/elee1766/jobshop/commit/6146655dc9d4310434b04cac9f3b14afc70786e4))

## [0.2.5](https://github.com/elee1766/jobshop/compare/v0.2.4...v0.2.5) (2026-11-06)


### Bug fixes

* **packs/github:** release-merge waits for a running release workflow ([#764](https://github.com/elee1766/jobshop/issues/764)) ([3ee42ce](https://github.com/elee1766/jobshop/commit/3ee42ce1b6b8e6a5ff260659b9b13270e51d911e))

## [0.2.4](https://github.com/elee1766/jobshop/compare/v0.2.3...v0.2.4) (2026-11-06)


### Features

* **packs/github:** hand cross-repo prs to the owning instance's pr-review (JOB-33) ([#760](https://github.com/elee1766/jobshop/issues/760)) ([16efd81](https://github.com/elee1766/jobshop/commit/16efd8111265abf686e22d4c4179a81a0370bb36))


### Bug fixes

* **control:** wait for both step writes in TestAutoStepHTTPRetryAfter ([#757](https://github.com/elee1766/jobshop/issues/757)) ([#758](https://github.com/elee1766/jobshop/issues/758)) ([ed1651f](https://github.com/elee1766/jobshop/commit/ed1651f27f43c0b2c0514828db8e8d72d85bf611))
* **store:** race storetest cases, postgres reads read committed (JOB-38) ([#754](https://github.com/elee1766/jobshop/issues/754)) ([a27348d](https://github.com/elee1766/jobshop/commit/a27348d96b1a4f9e7d5f32ce84ddfe1b82c11c9c))


### Performance

* **control:** list closed sessions from saved state, scan orphans in the pass (JOB-32) ([#747](https://github.com/elee1766/jobshop/issues/747)) ([698a31b](https://github.com/elee1766/jobshop/commit/698a31b81aaa0d501352484815f11fd6577f468e))

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
