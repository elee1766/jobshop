# Desk chat attachments

2026-11-07

Use an isolated controller with an image-capable `type: chat` operator pool.
Do not run mutation checks against a production controller.

- Start a named draft. Upload a text file, remove it, then use Paste document.
  Dismiss the draft. No chat or message should have been created.
- Paste a PNG screenshot into the composer. It should become a preview chip.
  Paste more than 4,000 characters or at least 40 lines. It should become a
  text-document chip without replacing the message text.
- Send both. The agent receives native ACP `image` and `resource_link` blocks.
  Text documents are redacted, saved in the session mount and read on demand,
  not expanded into the prompt. Linked files survive a controller restart.
- Ask an image-capable agent to read a picture with its image-reading tool.
  Its native image output should appear in the transcript. Reload and reopen
  the same chat. Both sent and received pictures should still display.
- PNG, JPEG and GIF are supported. SVG and MIME mismatches are rejected.
  A message permits eight files totaling 8 MiB. UTF-8 documents are at most
  2 MiB. Images are limited to 64 megapixels and 8 MiB per event.
- Reaping a closed session also removes its attachment files. Archiving alone
  does not remove them.

Automated checks against the isolated controller:

```sh
DESK_URL=http://127.0.0.1:7484 npm --prefix ui run test:e2e -- desk.spec.js
DESK_ATTACHMENTS=1 DESK_URL=http://127.0.0.1:7484 npm --prefix ui run test:e2e -- attachments.spec.js --workers=1
```

For the native output check, set `DESK_IMAGE_OUTPUT_DIR` to a local directory
visible to that controller's operator agent. The test writes its own PNG there,
asks the real agent to read it, verifies image output and replay, then removes
that file and archives only its own test chat. The configured agent must forward
native ACP image output. Jcode's ACP adapter must include SidePaneImages support.
No image-generation API or paid image-generation request is needed.
