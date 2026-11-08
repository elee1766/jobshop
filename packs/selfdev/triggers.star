# The controller calls this only once the learnings or edits gate holds.
def start(event, args):
    return [start_run(workflow = args["workflow"], key = args["key"])]
