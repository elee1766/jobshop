# Matrix posts retry on 429 and 5xx (#196)

After a deploy that includes #196, on gasserver: `journalctl -u
jobshop-gateway* --since -1d | grep -i 'matrix\|post'` shows no
dropped-reply WARN that a retry would have covered. If a 429 or 5xx
does occur, mautrix logs the retry and the reply still appears in the
thread.
