def _existing(run, id):
    for step in run.steps:
        if step.id == id:
            return True
    return False


def _outcome(run, id):
    for step in run.steps:
        if step.id == id:
            return step.outcome
    fail("missing recorded step: " + id)


def _step(run, id, pool, title, prompt, retries=2, retry_pool="", needs=[]):
    if _existing(run, id):
        return []
    return [start_step(
        id=id,
        pool=pool,
        title=title,
        prompt_file=prompt,
        inputs=run.inputs,
        retries=retries,
        retry_pool=retry_pool,
        needs=needs,
    )]


def finish(run, step):
    # Review recovery can reopen an earlier step while a completion is queued.
    for recorded in run.steps:
        if recorded.status != "closed" or recorded.resolution not in ["done", "skipped"]:
            return []
    return [finish_run(outcome="done")]


# Route outcome to the pool of the one implement step. Only the selected
# step is created; an unknown outcome fails the callback, which sends the
# run to a human.
_AUTHORS = {"scoped": "dev-sonnet", "primary": "dev"}


def _route(run, tracker, issue):
    return _step(run, "route", "triage", "Route " + issue, "prompts/" + tracker + "/route.md", retries=1)


def _implement(run, step, tracker, issue):
    if step.outcome not in _AUTHORS:
        fail("unknown route outcome: " + repr(step.outcome))
    # A failed attempt or a review resume moves the step to dev.
    return _step(run, "implement", _AUTHORS[step.outcome], "Implement " + issue,
                 "prompts/" + tracker + "/implement.md", retries=3, retry_pool="dev", needs=["route"])


def pr_work_start(run):
    return _route(run, run.inputs["tracker"], "#" + str(run.inputs["issue"]))


def pr_work_linear_start(run):
    return _route(run, "linear", str(run.inputs["issue"]))


def _pr_completed(run, step, tracker, issue):
    if step.id == "route":
        return _implement(run, step, tracker, issue)
    if step.id == "implement":
        if step.outcome != "pr":
            return finish(run, step)
        return _step(run, "review-merge", "refinery", "Review and merge PR for " + issue,
                     "prompts/" + tracker + "/merge.md", needs=["implement"])
    if step.id == "review-merge":
        if _outcome(run, "implement") != "pr":
            return finish(run, step)
        return _step(run, "post-merge", "deploy-check", "Check master CI and deploy for " + issue,
                     "prompts/" + tracker + "/post-merge.md", needs=["review-merge"])
    if step.id == "post-merge":
        return finish(run, step)
    fail("unexpected PR work step: " + step.id)


def pr_work_completed(run, step):
    return _pr_completed(run, step, run.inputs["tracker"], "#" + str(run.inputs["issue"]))


def pr_work_linear_completed(run, step):
    return _pr_completed(run, step, "linear", str(run.inputs["issue"]))


def pr_review_start(run):
    return _step(run, "review-merge", "refinery", "Review and merge PR #" + str(run.inputs["pr"]),
                 "prompts/pr-review/merge.md")


def pr_review_completed(run, step):
    if step.id == "review-merge":
        if step.outcome != "merged":
            return finish(run, step)
        return _step(run, "post-merge", "deploy-check", "Check master CI and deploy for PR #" + str(run.inputs["pr"]),
                     "prompts/pr-review/post-merge.md", needs=["review-merge"])
    if step.id == "post-merge":
        return finish(run, step)
    fail("unexpected PR review step: " + step.id)


def triage_start(run):
    return _step(run, "diagnose", "triage", "Triage " + run.inputs["signal"] + ": " + run.inputs["subject"],
                 "prompts/triage.md", retries=1)


def ci_start(run):
    return _step(run, "fix", "dev", "Fix red master CI at " + run.inputs["sha"], "prompts/ci-fix.md")


def ci_completed(run, step):
    if step.id == "fix":
        return _step(run, "merge", "refinery", "Merge CI fix for " + run.inputs["sha"],
                     "prompts/ci-merge.md", needs=["fix"])
    if step.id == "merge":
        return finish(run, step)
    fail("unexpected CI repair step: " + step.id)


def release_start(run):
    return _step(run, "merge", "refinery", "Merge release PR #" + str(run.inputs["pr"]),
                 "prompts/release-merge.md")


def drill_start(run):
    return _step(run, "fail", "dev", "Escalation drill for #" + str(run.inputs["issue"]),
                 "prompts/drill.md", retries=1)
