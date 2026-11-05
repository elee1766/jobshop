# Session socket bound once (#479, 2026-11-05)

The admin socket serves before startup reopens open sessions. A session
created in that window was opened twice: the second `listen` removed the
live cli.sock and raced the first bind ("address already in use" in
TestSessionScope). `openSession` now checks `c.sessions` and binds under
`c.mu`, so a session already open is left alone.
