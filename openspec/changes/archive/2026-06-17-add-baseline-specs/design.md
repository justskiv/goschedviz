## Context

goschedviz is an existing, working educational tool. This change adds no
behavior — it records the as-built architecture as the OpenSpec
baseline. The design below documents the decisions already embodied in
the code so future changes have a reference point. The system is a
clean, layered pipeline: a collector runs a target program under
`GODEBUG=schedtrace`, a parser turns trace lines into validated
snapshots, a state object keeps current values plus rolling history, and
a terminal presenter renders them.

## Goals / Non-Goals

**Goals:**
- Capture the current behavior of each capability accurately enough to
  review code against and to base future changes on.
- Keep capability boundaries aligned with the existing package layout so
  specs map cleanly to code.

**Non-Goals:**
- No code changes, refactoring, or behavior changes.
- Not specifying internal data structures, exact widget layout
  proportions, or visual styling — those are implementation detail.
- Not adding new metrics, flags, or output formats.

## Decisions

- **Collect via `GODEBUG=schedtrace` over a child process.** The target
  is compiled to a temporary binary and run with the schedtrace
  environment variable; metrics are read from its stderr. This uses the
  Go runtime's own instrumentation with no code changes to the target,
  at the cost of spawning a process and parsing text. Alternative
  (runtime/metrics or instrumenting the target) was rejected because it
  would not expose scheduler-internal counters the same way.
- **Goroutine count via an opt-in `PROCMETR` side channel.** schedtrace
  does not report goroutine totals, so the embeddable `pkg/metrics`
  reporter prints `PROCMETR num_goroutines=<n>` to stderr and the parser
  carries the last value into snapshots. This keeps the count optional —
  targets that do not import the reporter simply omit it.
- **Parser validates snapshots before emitting.** Inconsistent or
  all-zero lines are dropped so the UI never shows malformed data; this
  trades a few skipped frames for display correctness.
- **Two clocks: trace period vs. UI refresh.** The collector emits at
  the schedtrace period; the CLI refreshes the UI on its own fixed tick.
  Decoupling keeps rendering steady regardless of trace cadence.
- **Presenter behind an interface.** The UI is consumed through a small
  `Presenter` interface (Start/Stop/Update/Done), and the terminal layer
  hides the terminal API behind an interface so widgets can be tested
  with a mock terminal.

## Risks / Trade-offs

- [Baseline drift from code] → Validate with `openspec validate` and
  keep specs at the behavior level so small refactors don't invalidate
  them.
- [Text parsing is format-sensitive] → schedtrace output format could
  change across Go versions; the parser's strict regex and validation
  make mismatches fail closed (no snapshot) rather than show garbage.
- [Documentation-only scope misread as a feature] → All requirements are
  ADDED and describe current behavior; this is explicit in the proposal.

## Open Questions

- None. This baseline reflects the code as currently implemented.
