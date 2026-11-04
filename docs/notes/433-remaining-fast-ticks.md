# #433 last 50ms test tickers

2026-11-04. The tests that still ran a controller on a real 50ms tick
now tick hourly and drive the supervisor with passes: runRoot (about 30
run, batch and escalation tests), TestMetrics, TestTraces and
TestSessionsEndedMetric. The last three also use in-process hosts, since
agent exits reach the controller in-process since #467. On gasserver
(6-core cgroup quota, shared) `go test -race ./pkg/control` went from
41s to 28s. The suite is now mostly latency bound: the remaining
real-time waits are host and agent startup, poll intervals and
TestChatPoolKeys's 3s idle timeout.
