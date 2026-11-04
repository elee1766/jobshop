# Verify: provider-owned resource slots (#407)

On gasserver, with a provider resource in `jobshop.yaml`:

```yaml
resources:
  emu: {provider: {command: "sh /tmp/emu-prov.sh"}}
```

where `/tmp/emu-prov.sh` runs `echo "$@" >> /tmp/emu-prov.log`, and a
test workflow whose step has `needs_resources: [emu]`:

1. `jobshop config reload`; it succeeds.
2. `jobshop run start <wf> --var issue=1 --key a`. `jobshop ready` does
   not list the step: no slot. `/tmp/emu-prov.log` gets a `demand` line.
3. `jobshop resource add emu e1 --env ANDROID_SERIAL=emulator-5554`. The
   step is ready. Once claimed, `jobshop resource list` shows e1 `held`
   with the step's ticket, and the session's prompt starts with
   `export ANDROID_SERIAL='emulator-5554'`.
4. Close the step. Within a pass the log has one `reset e1` line and
   `resource list` shows e1 `ready`.
5. Start a second run, let it claim, then `jobshop resource unhealthy emu
   e1 -r test`. The step is open again with a "went unhealthy" comment
   and `failures` unchanged. e1 is `gone`.
6. `jobshop resource remove emu e1`: e1 is no longer listed.
