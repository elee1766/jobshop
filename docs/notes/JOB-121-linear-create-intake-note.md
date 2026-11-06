# linear create: intake note (JOB-121)

2026-11-06

`jobshop linear create` without `--state` leaves the issue in the team's
default state (Backlog for JOB). Linear intake (`linear_issue_no_run`)
only starts runs for issues in `linear_intake_states` (default Todo), so
such an issue sat with no run until someone moved it.

The command keeps Linear's default. After a create without `--state` it
asks the controller (`linear.intake`, open to every caller, reads config
only) which states the team's `linear_issue_no_run` signals take. When
a signal watches the team and the new issue's state is not one of them,
it prints one line on stderr:

    note: JOB-9 is in Backlog; Linear intake only takes Todo (linear_intake_states), so no run starts until it moves there (jobshop linear state JOB-9 "Todo")

No note when no signal watches the team, when a signal takes any state,
or when the controller does not know the method. The JSON result now
carries `state`, the state the issue landed in.
