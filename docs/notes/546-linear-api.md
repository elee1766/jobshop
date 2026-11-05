# jobshop linear api (#546)

Added 2026-11-05.

```
jobshop linear api [--team ENG] [-f name=value]... [-F name=json]... [--input FILE|-] [QUERY]
```

- Query from the argument, `-f query=...` or `--input` (file or `-`).
  An input whose first non-space byte is `{` followed by a `"` key is a
  JSON request (`query`, `variables`); anything else is bare GraphQL.
- `-f` sets a string variable, `-F` a JSON one; `@file` reads the value.
- Auth: `LINEAR_API_KEY` if set (no controller needed for the token).
  Otherwise the team identity (`identity.token`) of `--team` or of the
  first issue key found in the variables (names in sorted order). No
  key and no team is an error; there is no default identity.
- Mutations ask the guard first, action `linear-api-mutation`, target
  the team, detail the operation name and top-level fields. Deny, ask
  or no controller: not sent. This applies with an own key too.
  `subscription` is refused.
- Every call is recorded in the activity feed as
  `linear api <query|mutation> <Name>`; no variables, query text or
  token.
- Stdout is the response body as is. Exit 1 when it has `errors`
  (printed to stderr too) or the HTTP status is not 200.
- The operation check is a token scan, not a GraphQL parser: it skips
  comments, strings and leading fragments, then reads the first
  operation keyword.
