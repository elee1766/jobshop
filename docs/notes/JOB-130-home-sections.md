# One home page for waiting and in flight at every width

2026-11-08. JOB-130.

Waiting, not-moving and in flight are sections of Desk's home page at every
width. Before, the phone tabs hid the others (`data-phone` CSS), so in flight
was a separate page on mobile and resizing switched the visible content.

- The CSS no longer hides home sections by tab. `waiting` and `in flight` in
  the phone nav scroll `.work` to `#waiting` or `#inflight` and put that id in
  the URL hash, so back and forward, reload and shared links land on the same
  section. The main `home` button and `catch-up` scroll to the top.
- `chats` and `more` stay separate destinations: chat is its own column on
  desktop, `more` is the page menu that desktop shows in the header.
- Filter, sort and collapse state live in the one home component, so a
  resize or tab press never resets them. Leaving for another view and
  returning keeps them while Desk stays loaded.
- The hash is ignored on other views and only `waiting` and `inflight` are
  accepted.

`ui/e2e/home-sections.spec.js` covers resizing both ways, the shortcuts,
history and direct links (see `docs/verify/home-sections.md`).
