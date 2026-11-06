# Prefixed dev_pool default (#880)

2026-11-06. A string pack default that holds a template is rendered at
load over the given vars, the other defaults and `pack_prefix`, so
packs/github's `dev_pool: "{{.pack_prefix}}dev"` gives a prefixed
instance (`acphost-`) `acphost-dev`. Before, the default stayed `dev`
and a `acphost-pr-work` run started without `--var dev_pool` put its
implement step in the unprefixed instance's dev pool (acphost #210,
#211). A var the includer sets is used as given. A templated default
that reads another templated default sees its template text.

Run start also rejects a step pool filled from a run var when another
packs entry declares that pool and the workflow's own entry does not
(`pool "dev" belongs to pack instance ...`). Literal step pools and host
pools are not checked.
