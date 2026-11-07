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

def completed(run, step):
    if step.id not in ["curate", "evaluate"]:
        fail("unexpected completed step: " + step.id)
    if step.status != "closed" or step.resolution not in ["done", "skipped"]:
        fail("step is not finished: " + step.id)
    return [finish_run(outcome = "done")]
