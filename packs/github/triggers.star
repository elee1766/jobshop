load("lib.star", "green", "listed", "older_than")

# Trigger handlers. args carry the pack vars each one needs; they return
# actions and never write to GitHub themselves.

# replan files one planning task when dev has nothing ready. The task key
# dedups while it is open; the trigger cooldown runs from its close.
def replan(event, args):
    ready = tickets(pool = args["dev_pool"], status = "open", limit = 1)
    if ready:
        return []
    return [create_ticket(
        title = "Plan next step",
        key = "replan",
        pool = args["plan_pool"],
        priority = 1,
        labels = ["plan"],
    )]

# master_red starts ci-repair for the newest completed run of the CI
# workflow on the CI branch when it failed, keyed by its head SHA.
def master_red(event, args):
    runs = github.get(
        "/repos/" + args["repo"] + "/actions/runs",
        params = {"branch": args["ci_branch"], "status": "completed", "per_page": "30"},
    ).json()["workflow_runs"]
    runs = [r for r in runs if r["name"] == args["ci_workflow"]]
    if not runs or runs[0]["conclusion"] != "failure":
        return []
    sha = runs[0]["head_sha"]
    return [start_run(workflow = "ci-repair", key = sha, inputs = {"sha": sha})]

# release_pr starts release-merge for the lowest-numbered open release PR
# whose checks all passed.
def release_pr(event, args):
    label = args["release_pr_label"]
    if not label:
        return []
    prs = [pr for pr in _open_prs(args["repo"]) if not pr["isDraft"] and label.lower() in pr["labels"] and green(pr)]
    if not prs:
        return []
    n = sorted([pr["number"] for pr in prs])[0]
    return [start_run(workflow = "release-merge", key = "release-pr-" + str(n), inputs = {"pr": str(n)})]

# Scan signals: each returns {subject: detail} for the controller, which
# keeps per-subject history, cooldowns and caps.

# issue_no_run: open issues from the human, listed authors or bots, or
# labeled accepted, older than 30 minutes and not planning items. The
# controller drops issues that already have an open run.
def issue_no_run(args):
    ok = listed(args["human"] + "," + args["authors"])
    skip = ["target", "needs-decision", "needs-human", "tracking", "proposed"]
    out = {}
    issues = github.get("/repos/" + args["repo"] + "/issues", params = {"state": "open", "per_page": "100"}, pages = 2).json()
    for issue in issues:
        if "pull_request" in issue:
            continue
        labels = [label["name"] for label in issue["labels"]]
        author = issue["user"]
        if not (author["type"] == "Bot" or author["login"].lower() in ok or "accepted" in labels):
            continue
        if [label for label in labels if label in skip]:
            continue
        if issue["title"].lower().startswith("drill"):
            continue
        if not older_than(issue["created_at"], "30m"):
            continue
        out[str(issue["number"])] = "open, no run"
    return out

# no_merge: open PRs but no commit on the CI branch for 6 hours.
def no_merge(args):
    if not github.get("/repos/" + args["repo"] + "/pulls", params = {"state": "open", "per_page": "1"}).json():
        return {}
    commit = github.get("/repos/" + args["repo"] + "/commits/" + args["ci_branch"]).json()
    date = commit["commit"]["committer"]["date"]
    if not older_than(date, "6h"):
        return {}
    return {commit["sha"][:12]: "last commit " + date}

# pr_unmerged: green, non-draft PRs untouched for 2 hours, other than the
# human's, skipped authors and labels, and release PRs.
def pr_unmerged(args):
    skip_authors = listed(args["human"] + "," + args["skip_authors"])
    skip_labels = listed(args["skip_labels"])
    release = args["release_pr_label"].lower()
    out = {}
    for pr in _open_prs(args["repo"]):
        if pr["isDraft"] or pr["author"] in skip_authors:
            continue
        if [label for label in pr["labels"] if label in skip_labels]:
            continue
        if release and release in pr["labels"]:
            continue
        if not green(pr) or not older_than(pr["updatedAt"], "2h"):
            continue
        out[str(pr["number"])] = "green, unmerged since " + pr["updatedAt"]
    return out

_OPEN_PRS = """
query($owner: String!, $name: String!) {
  repository(owner: $owner, name: $name) {
    pullRequests(states: OPEN, first: 100) {
      nodes {
        number isDraft updatedAt
        author { login }
        labels(first: 20) { nodes { name } }
        commits(last: 1) { nodes { commit { statusCheckRollup { contexts(first: 100) { nodes {
          __typename
          ... on CheckRun { conclusion }
          ... on StatusContext { state }
        } } } } } }
      }
    }
  }
}
"""

# _open_prs is one GraphQL read of open PRs with lowercase labels, the
# author login and the head commit's check conclusions.
def _open_prs(repo):
    owner, name = repo.split("/")
    data = github.graphql(_OPEN_PRS, variables = {"owner": owner, "name": name}, repo = repo)
    out = []
    for pr in data["repository"]["pullRequests"]["nodes"]:
        checks = []
        for commit in pr["commits"]["nodes"]:
            rollup = commit["commit"]["statusCheckRollup"]
            if rollup:
                for c in rollup["contexts"]["nodes"]:
                    checks.append(c.get("conclusion") or c.get("state") or "")
        out.append({
            "number": pr["number"],
            "isDraft": pr["isDraft"],
            "updatedAt": pr["updatedAt"],
            "author": (pr["author"] or {}).get("login", "").lower(),
            "labels": [label["name"].lower() for label in pr["labels"]["nodes"]],
            "checks": checks,
        })
    return out
