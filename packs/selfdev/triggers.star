def handle(event, args):
    # The controller admits timers only after the learnings or edits gate passes.
    if event.source != "timer" or event.event != "timer":
        return []
    return [start_run(workflow = args["workflow"], key = args["key"])]
