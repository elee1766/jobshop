# Functions are deliberately not called at module load time.
def http_probe(url):
    response = http.get(url, timeout = "2s")
    if response.status_code != 200:
        fail("probe returned non-200 status")
    return response.json()

def shell_probe():
    result = shell(["/bin/printf", "%s", "ok"], timeout = "2s")
    if result.exit_code != 0:
        fail("probe command failed")
    return {"stdout": result.stdout, "exit_code": result.exit_code}

def decide(state):
    result = jev(
        instructions = "Choose the implementation route from these options.",
        options = {
            "scoped": "Specified implementation with limited scope",
            "primary": "Investigation or higher-risk implementation",
        },
        state = state,
        default = "primary",
        timeout = "2s",
    )
    return {"choice": result.choice, "defaulted": result.defaulted}
