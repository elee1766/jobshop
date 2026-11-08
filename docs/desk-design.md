# Desk design and UX

2026-11-08. Agent guidance for new work and incremental changes. This describes
the intended direction, not a claim that every existing view already meets it.

## Direction

Desk's theme is based on Plan 9/Acme. Its interaction philosophy draws on Unix:
simple, explicit tools that work together, with the features needed to do the
job. Keep information readable and actions close to their context.

This is a working interface, not a historical replica. Do not require Acme
knowledge, mouse chords or hidden gestures. Simplicity must preserve
accessibility, useful features, error handling and clear feedback.

## Visual language

- Use flat panes, compact text bars, square controls and thin rules to establish
  hierarchy. Avoid decorative cards, shadows, gradients and excessive nesting.
- The light theme uses pale paper, cyan tag bars, yellow selection and restrained
  borders. Use the existing dark theme rather than hardcoding light colors.
- `ui/shell/tokens.css` is the source of truth for colors, typography, spacing and
  control dimensions. `ui/shell/app.css` holds shared element and layout styles.
  Reuse semantic tokens. Add a shared token only for a real recurring need.
- Align text baselines, visible text insets and adjacent controls. Bars serving
  the same role should share height and padding. A matching outer box alone does
  not establish alignment.
- Use body type for prose and readable titles. Use monospace where it helps scan
  structured data, identifiers or code. Do not shrink text to fit excess metadata.

## Information hierarchy

- Lead with the readable title or task, not an opaque ID. Keep IDs available in
  details, copying and exact-reference tooltips without crowding overview rows.
- Show enough information to understand state and choose the next action at a
  glance. Group related items and consolidate secondary metadata onto compact
  rows. Do not present every field as a separate line or dump raw objects.
- Put raw metadata and verbose context behind labeled disclosures. Keep important
  errors, outcomes and actions visible. Default views should emphasize relevant
  work, with explicit ways to inspect filtered or historical information.
- Do not repeat the same fact in adjacent forms, such as `O open`. A compact
  status marker needs an accessible full label, not a second visible status.
  Never rely on color alone to communicate state.
- Summaries must reflect actual data. Do not invent progress, metrics or results.

## Copy and interaction

- Use short, lowercase navigation and control labels: `home`, `trace`, `learn`,
  `sys`, `catch-up`, `send`, `attach`, `paste`. Prefer one word when it remains
  clear. Use more words when needed to distinguish actions safely.
- Preserve authored text, proper names, identifiers and code. Do not lowercase
  data with CSS or rewrite user content to match the control-label convention.
- Use links for navigation and buttons for actions. Links are underlined with a
  restrained background and visible hover state. Clickable controls need an
  appropriate cursor, visible focus and disabled behavior.
- Keep common actions visible and contextual. Use disclosures and small popovers
  for secondary choices rather than large dialogs for small tasks. Destructive
  actions still need appropriate confirmation or a safe recovery path.
- Expose exact references and times through tooltips or details, with keyboard
  and touch access. Hover is supplemental, never the only route to information
  or an action.
- Show pending, queued, working, disconnected, failed and completed states where
  relevant. A click must receive timely feedback, not leave users guessing or
  inviting duplicate submissions. Do not present queued work as completed.

## Layout and accessibility

- Use space for the work itself. Keep tag bars and inputs proportionate rather
  than stretching small controls or surrounding them with explanatory filler.
- Keep related panes aligned and independently usable. On narrow screens, turn
  secondary inspectors into dismissible views without trapping navigation.
- Preserve semantic headings, labels, native controls, keyboard operation and
  visible focus. Compact visual density must not sacrifice readable text,
  adequate pointer targets or contrast.
- Long titles, links, code and populated tables must not push controls outside
  the viewport. Do not silently hide needed actions to make a screenshot fit.
- Preserve reading position during live updates and history loading. Bound
  rendered work for long transcripts without breaking selection, disclosures,
  accessible navigation or access to older content.

## Before shipping

Inspect representative populated views at desktop and mobile sizes, in light
and dark themes. Check long content, empty and error states, pending work,
keyboard focus, touch access, bar alignment and overflow. Test the affected
workflow, not just the static page or an empty fixture.

Use the relevant [browser acceptance runbooks](verify/) and
[Desk live delivery checks](verify/desk-redesign.md). Mutating checks belong on
an isolated controller. Changes to this guide do not authorize unrelated visual
rewrites or deployment changes.
