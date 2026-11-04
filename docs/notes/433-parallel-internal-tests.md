# pkg/control internal tests run in parallel (#433)

2026-11-04. The `*_internal_test.go` files and `subscribe_test.go` had no
`t.Parallel()`, so ~110 tests (about 14s of test time under -race) ran
one after another after the parallel tests finished. Each builds its own
controller, store and temp dir and none touch process state, so they now
call `t.Parallel()`. `respawn_env_internal_test.go` stays serial: it uses
`t.Setenv`.

`go test -race ./pkg/control` on gasserver: 31s to 21s wall.
