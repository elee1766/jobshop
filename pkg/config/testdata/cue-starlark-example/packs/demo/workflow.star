def begin(run):
    return [start_step(
        id = "implement",
        pool = run.inputs["pool"],
        prompt_file = "implement.md",
        inputs = {"task": run.inputs["task"]},
        title = "Implement: " + run.inputs["task"],
        retries = 1,
        output_schema = {
            "type": "object",
            "required": ["summary"],
            "properties": {"summary": {"type": "string"}},
            "additionalProperties": False,
        },
    )]

def completed(run, step):
    if step.id != "implement":
        fail("unexpected completed step: " + step.id)
    outcome = step.outcome if step.outcome else step.resolution
    if not outcome:
        fail("completed step has no recorded outcome or resolution")
    return [finish_run(outcome = outcome, output = step.output)]
