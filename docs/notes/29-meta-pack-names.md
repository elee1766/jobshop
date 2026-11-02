# Meta pack names (#29)

Top-level `meta` keys on `ticket create`/`update` must be `github`,
`trigger`, or listed in `meta_packs:` (a string list in jobshop.yaml,
picked up on reload). Unknown names fail with `invalid`. The store only
checks the `{"<pack>": {...}}` shape.
