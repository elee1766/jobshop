# Session ids carry a kind tag (#615)

2026-11-05. New session ids are `<prefix>-s-<sonyflake>`, e.g.
`js-s-00337y40kc002`; every other kind keeps `<prefix>-<sonyflake>`. One
generator and epoch, so ids still sort by time within a kind. The store's
`insert` picks the tag from the kind (`ids.Next(prefix, tag)`).

Stored ids are not rewritten. Session directories and journals use the id
as is, so old `js-...` sessions keep their paths, and every lookup goes
through the store by id, so they still attach, list and reap.
`ids.IsSession(id)` is true only for tagged ids: code that uses it to skip
a lookup must fall back to the store when it is false.
