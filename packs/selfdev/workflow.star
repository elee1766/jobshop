def curate(run):
    return [start_step(
        id = "curate",
        pool = "learn",
        retries = 1,
        title = "Curate learnings",
        prompt_file = "prompts/curate.md",
    )]

def evaluate(run):
    return [start_step(
        id = "evaluate",
        pool = "learn",
        retries = 1,
        title = "Evaluate memory edits",
        prompt_file = "prompts/evaluate.md",
    )]

# Live check (#635): run a docs/verify runbook's readonly blocks in a
# worktree of master, in the dev pool sandbox, and comment the report on
# the issue. A failing block fails the step and the run escalates to plan.
def live_check(run):
    return [start_step(
        id = "check",
        pool = "dev",
        script = check_runbook,
        worktree = True,
        timeout = "20m",
        retries = 0,
        title = "Live check " + run.inputs["runbook"],
        inputs = {"runbook": run.inputs["runbook"], "issue": str(run.inputs["issue"])},
    )]

def check_runbook(run, step):
    # JOBSHOP_BIN is the controller's binary; units have no jobshop on PATH (#669).
    report = shell(["sh", "-c", "exec \"${JOBSHOP_BIN:-jobshop}\" runbook run \"$0\" 2>&1", step.inputs["runbook"]])
    posted = shell(["gh", "issue", "comment", step.inputs["issue"], "--repo", "elee1766/jobshop", "--body", report.stdout])
    print(report.stdout)
    if posted.exit_code != 0:
        fail("gh issue comment failed: " + posted.stderr)
    if report.exit_code != 0:
        fail("a readonly runbook block failed; report posted on #" + step.inputs["issue"])
    return result(outcome = "pass")

def completed(run, step):
    if step.id not in ["curate", "evaluate", "check"]:
        fail("unexpected completed step: " + step.id)
    if step.status != "closed" or step.resolution not in ["done", "skipped"]:
        fail("step is not finished: " + step.id)
    return [finish_run(outcome = "done")]
