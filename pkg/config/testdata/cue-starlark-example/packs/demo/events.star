def handle(event, args):
    if event.source == "timer":
        return []
    tasks = event.payload.get("tasks")
    if type(tasks) != "list":
        fail("event.payload.tasks must be a list")
    actions = []
    for task in tasks:
        if type(task) != "dict":
            fail("each task must be a dictionary")
        key = task.get("key")
        text = task.get("task")
        if type(key) != "string" or not key:
            fail("task.key must be a nonempty string")
        if type(text) != "string" or not text:
            fail("task.task must be a nonempty string")
        actions.append(start_run(
            workflow = args["workflow"],
            key = "demo:" + key,
            inputs = {"task": text, "pool": args["pool"]},
            title = "Implement: " + text,
        ))
    return result(actions=actions, output={"selected": len(actions)})
