# Learning nudge (#271)

2026-11-04. Item "Nudge" of target #247 B.

- Prompt rule `learn-nudge` in packs/github, `at: after`, no match: it
  follows the pool prompt on every ticket turn of every pool on the host
  (dev, refinery, deploy-check, plan, stall, and the operator, since
  prompt rules are host-wide). Text: before `jobshop close` or
  `release`, record a correction, failure or workaround with
  `jobshop learn ... --fix`; nothing for routine success.
- Chat sessions are not excluded: `after` rules go where the pool prompt
  goes, so an operator chat gets the nudge on its first message turn and
  not on later ones. Operator improvement sessions close runs and are
  where most corrections happen.
- With `operator-improve` (also `at: after`, operator `improve:*` keys)
  both apply in config order: packs/github's rules before packs/operator's.
- A prefixed packs/github instance adds its own copy (`<prefix>learn-nudge`),
  so a host with two instances shows the text twice. Same as `recall`.
