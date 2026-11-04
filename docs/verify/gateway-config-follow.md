# Gateway picks up sender changes without a restart (#307)

1. Note the gateway start time: `systemctl show -p ActiveEnterTimestamp jobshop-gateway-local`.
2. Add a Matrix user to `gateway.local.senders` (NixOS: edit the host
   config and switch; elsewhere edit jobshop.yaml and run
   `jobshop config reload`).
3. Within 30s `journalctl -u jobshop-gateway-local` shows
   `gateway: config applied` with the new sender count.
4. From that user, post in an allowed room. The gateway reacts 👀 and the
   session replies in the thread.
5. The start time from step 1 is unchanged (no restart).
6. Remove the sender again and reload; after 30s a message from them is
   ignored.
