# Desk trace and chats

2026-11-07, #966.

Trace shows the connected ticket graph, including closed tickets, dependency edges, parent relationships and recorded creation genealogy. The controller returns at most 2,000 nodes and reports truncation. Selecting a node opens readable details, relationships, tracker links, comments and optional raw data. State updates arrive through the existing live projection.

Markdown preserves explicit HTTPS links. `jobshop://ticket/<id>` and `jobshop://session/<id>` links navigate within the current Desk instance. Plain identifiers are not automatically linked. Agents receive the same link guidance, and attention context retains the complete response rather than the short preview.

Chats use a recent-first searchable dropdown. A named new-chat draft remains local until its first message is sent. Dismissing it restores the previous chat's composer and attached work references.
