# Planner proposals live check (#488)

On gasserver, after the deploy carrying #488:

1. Wait for a state with no unchecked target item that the planner can
   cover and a quiet dev pool (`jobshop session list`, fewer than 3 open
   issues with a running pr-work run).
2. After the next planning pass, `gh issue list --label proposed` shows
   exactly one issue titled `Proposed target: ...`, each item citing a
   source, and a comment mentioning the human with approve, edit or close.
3. `jobshop attention` lists a `decision` item for it.
4. After a second pass, still one `proposed` issue.
5. Relabel it `target` (`gh issue edit <n> --remove-label proposed
   --add-label target`). The following pass files issues for its items
   and closes the attention ticket.
