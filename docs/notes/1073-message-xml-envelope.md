# Messages as XML envelopes (#1073, 2026-11-08)

Replaces the plain-text `Messages for you:` prompt described in
`16-messages.md`. Delivery is otherwise as before: claim, close
`delivered`, at most once, one prompt per delivery.

The model now reads

```
<messages>
<message from="admin" id="js-m1">title
body
Attachment 0: "a.png" (image/png, 12 bytes).</message>
<message from="controller" id="js-m2">next</message>
</messages>
```

- Built in `pkg/control/messagesxml.go`, used by `deliver()` in
  `messages.go`. Sender, id and text go through `encoding/xml`'s
  `EscapeText`, so `</message>`, `&`, `<`, `>` and quotes in user input
  cannot end an element or attribute. Newlines and tabs are written
  literally (EscapeText makes them `&#xA;`/`&#x9;`, which are put back).
  A user's own `&#xA;` stays `&amp;#xA;`.
- Order follows the claim order. The text is the title, the body on the
  next line, then the attachment lines, as before.
- Context-only messages add no element. If every message is context only,
  the turn has no message part, as before.
- The `session.prompt` preview shows `<messages>\n(next messages)\n</messages>`.
- Not changed: the original message text Desk shows
  (`PromptPresentation.Messages` and `Context` carry the text, not the
  envelope), controller-message handling, the first-turn pool prompt for
  chats, and the raw prompt reported in `turn_start`.
- Agents and packs that match the old `Messages for you:` header need the
  new form. The operator pack prompt is updated. A jobcity host config
  that names the old header is not touched here: coordinate with the GFX
  coordinator before switching or deploying it.
