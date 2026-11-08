# The controller calls this only once the learnings or edits gate holds.
# args.title names the job for Desk and the trace.
def start(event, args):
    return [start_run(workflow = args["workflow"], key = args["key"], title = args.get("title", ""))]
