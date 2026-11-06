# Tool labels and session agent (2026-11-06)

Desk showed every jcode tool call as "Running shell command" or "Reading
file": jcode's ACP adapter titles a call by tool type, and that title was
all Desk had.

- `session.list` / `SessionInfo` has `agent: {name, version}` from the
  agent's ACP `initialize` `agentInfo`, stored on the session ticket as
  `agent_name` / `agent_version` when a host connects.
- `ToolSummary` (session events) has `target` and `intent`; `SessionInfo`
  has `tool_target` and `tool_intent` for the running call. `target` is
  the first of `command`, `cmd`, `file_path`, `path`, `filePath`, `url`,
  `query`, `pattern` in the call's `rawInput`, else its first
  `locations` path. `intent` is `rawInput.intent` (jcode sends one per
  call). Both are one line, at most 200 characters, carried across the
  call's updates and redacted like tool input.
- Desk's `toolLabel(agent, tool)` (`ui/kit/format.js`): for `jcode` the
  intent leads with the command or path as detail; for any other agent
  the agent's title leads with the target as detail. The chat tool row,
  the working line and the holder chip use it.
