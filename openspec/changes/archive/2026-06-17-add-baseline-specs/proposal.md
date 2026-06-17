## Why

goschedviz is already implemented but has no specifications. Without a
written baseline, behavior lives only in code and is hard to reason
about, review against, or change deliberately. This change documents the
existing system as the OpenSpec baseline so future work has a contract
to build on.

## What Changes

- Document the as-built behavior of the whole tool as baseline specs.
  No code changes — every requirement describes current behavior and is
  marked **ADDED**.
- Capture four capabilities that together make up goschedviz: scheduler
  monitoring (collect + parse + state), the terminal dashboard, the
  embeddable goroutine reporter, and the CLI.
- On archive, these delta specs become the initial contents of
  `openspec/specs/`.

## Capabilities

### New Capabilities
- `scheduler-monitoring`: builds and runs a target `.go` program with
  `GODEBUG=schedtrace`, parses `SCHED` and `PROCMETR` lines into
  validated snapshots, and maintains current state plus rolling history.
- `terminal-dashboard`: renders snapshots and history as a terminal UI
  (current-values table, LRQ bar chart, gauges, dual history plots,
  legend, info box) and handles quit/resize events.
- `goroutine-metrics-reporter`: an embeddable reporter (`pkg/metrics`)
  that target programs import to emit live goroutine counts on stderr
  using the `PROCMETR` prefix.
- `cli`: the command-line entry point that parses flags, wires collector
  → state → presenter, drives the refresh loop, and shuts down
  gracefully.

### Modified Capabilities
<!-- None. openspec/specs/ is empty; this is the first baseline. -->

## Impact

- Documents (no behavior change): `cmd/goschedviz/`,
  `internal/collector/` (+`godebug/`), `internal/domain/`,
  `internal/ui/` (+`termui/` and widgets), `pkg/metrics/`.
- Dependencies referenced: `github.com/gizak/termui/v3`,
  `github.com/stretchr/testify` (tests). Go toolchain `go build` is
  invoked at runtime to compile the target.
- After archive: populates `openspec/specs/` with four capability specs.
