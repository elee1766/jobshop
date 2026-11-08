# push denies a push with no target and one to a protected branch in any
# ref form (origin/master, HEAD:master, refs/heads/main).
def push(event, args):
    target = event.get("target", "")
    if not target:
        return {"decision": "deny", "reason": "name the branch: jobshop guard push --target <branch>"}
    branch = target.replace(":", "/").split("/")[-1]
    if branch in args["protected"]:
        return {"decision": "deny", "reason": "operator sessions never push to " + branch + "; push a branch and open a PR"}
    return {}
