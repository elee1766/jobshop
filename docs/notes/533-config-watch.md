# Config watch (#533)

2026-11-05.

`reload: {watch: true, debounce: 2s}` in `jobshop.yaml` (off by default)
makes the controller watch its config with fsnotify and reload on change.

- **Watched:** `jobshop.yaml`, each `include:` file, every resolved pack
  dir recursively (extends included), and prompt files outside those dirs.
  `config.LoadFile` records them in `Config.Watch`. Files are watched
  through their parent dir so an editor's rename-over is seen.
- **Ignored:** editor temp files (`.*.swp`, `.swx`, `*~`, `.#*`, `4913`)
  and chmod-only events.
- **Debounce:** on the controller clock, restarted by each change; then
  the same path as `jobshop reload`. Success logs `config reloaded
  (watch)` with the changed files. Failure keeps the running config and
  raises attention kind `config-invalid`; any good reload clears it.
- **Watch set** is recomputed after every good reload (packs can be added
  or dropped), and the watcher starts or stops when `reload.watch` flips.
- `jobshop config show` has `watch: {on, files, dirs}` when on.
- NixOS: `services.jobshop.reloadOnChange` sets `reload.watch` and drops
  the `jobshop-config.path` unit from #531, so there is one reloader.
