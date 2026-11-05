-- Postgres replacements for sqlite queries that use engine features
-- (#493). Each takes the parameters of the sqlcdb query of the same
-- name, in the same order, as $1..$n, and returns the same columns.

-- name: RecallSearch
-- $1 is a to_tsquery expression from recallMatch. Lower score is
-- better, as with bm25. ts_rank's own length normalization counts
-- lexemes, which drops stopwords; bm25 counts every word, so the rank is
-- divided by the log of the body's word count instead: a short focused
-- match beats a long one.
SELECT kind, ref, ticket_id, project, CAST(at AS BIGINT) AS at_ms,
       ts_headline('english', body, q,
         'StartSel=[, StopSel=], MaxWords=16, MinWords=4, FragmentDelimiter=...') AS snippet,
       CAST(-ts_rank(tsv, q) / ln(2 + cardinality(regexp_split_to_array(trim(body), '\s+')))
         AS DOUBLE PRECISION) AS score
FROM recall, to_tsquery('english', $1) q
WHERE tsv @@ q
  AND ($2 = '' OR project = $2 OR (kind = 'learning' AND project = ''))
  AND ($3 = '' OR kind = $3)
  AND ($4 = '' OR ticket_id = $4)
ORDER BY score, id
LIMIT CAST($5 AS BIGINT)
