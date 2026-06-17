## ADDED Requirements

### Requirement: Target program validation

The collector SHALL validate its configuration before launching the
target and return an error without starting any process when the
configuration is invalid.

#### Scenario: Reject non-positive trace period
- **WHEN** the collector is started with a trace period less than or equal to zero
- **THEN** it returns a configuration error and does not launch the target

#### Scenario: Reject a target that is not a Go source file
- **WHEN** the target path is empty, does not exist, is a directory, or does not end in `.go`
- **THEN** the collector returns a configuration error and does not launch the target

### Requirement: Build and run the target with schedtrace

The collector SHALL compile the target Go program to a temporary binary
and run that binary with the environment variable
`GODEBUG=schedtrace=<period>`, where `<period>` is the configured period
in milliseconds. The target's standard input SHALL be connected so it
can run interactively.

#### Scenario: Target launched under schedtrace
- **WHEN** a valid target is started
- **THEN** the collector builds it and runs the resulting binary with `GODEBUG=schedtrace=<period>` set in its environment

#### Scenario: Build failure surfaces as an error
- **WHEN** the target fails to compile
- **THEN** the collector returns a build error and emits no snapshots

### Requirement: Parse scheduler trace lines

The parser SHALL recognize `SCHED <time>ms: ...` trace lines and extract
TimeMs, GoMaxProcs, IdleProcs, Threads, SpinningThreads, NeedSpinning,
IdleThreads, the global RunQueue length, and the per-P local run queue
lengths (with their sum). Lines that do not match the trace format SHALL
be ignored.

#### Scenario: Valid trace line is parsed
- **WHEN** a line matching the `SCHED ...` format with a bracketed list of per-P queue lengths is read
- **THEN** the parser produces a snapshot populated with all scheduler fields and the sum of the local run queues

#### Scenario: Unrelated output is ignored
- **WHEN** a stderr line does not match the trace or metrics formats
- **THEN** the parser produces no snapshot and continues reading

### Requirement: Parse goroutine metrics lines

The parser SHALL recognize lines prefixed with `PROCMETR` carrying a
`num_goroutines=<n>` value and SHALL carry the most recent goroutine
count into subsequently produced snapshots.

#### Scenario: Goroutine count attached to next snapshot
- **WHEN** a `PROCMETR num_goroutines=<n>` line is read and a later valid `SCHED` line is parsed
- **THEN** the produced snapshot reports the goroutine count `<n>` from the most recent metrics line

### Requirement: Validate parsed snapshots

The parser SHALL reject snapshots whose values are internally
inconsistent so that invalid trace output does not reach the UI. A
snapshot SHALL be rejected when GoMaxProcs is not positive, when
IdleProcs exceeds GoMaxProcs, when the number of per-P queues does not
equal GoMaxProcs, when Threads is less than SpinningThreads or less than
IdleThreads, or when any local run queue length is negative.

#### Scenario: Inconsistent snapshot rejected
- **WHEN** a parsed line violates any consistency rule (for example the per-P queue count differs from GoMaxProcs)
- **THEN** the parser discards it and produces no snapshot

#### Scenario: Suspect all-zero trace rejected
- **WHEN** the global and local run queues, idle processors, and spinning threads are all zero
- **THEN** the parser treats the line as suspect and produces no snapshot

### Requirement: Emit snapshots over a channel

Starting the collector SHALL return a receive-only channel of scheduler
snapshots, over which each validated snapshot is delivered in order.

#### Scenario: Validated snapshots are delivered
- **WHEN** the target emits valid trace lines
- **THEN** the corresponding snapshots are sent on the returned channel in the order they were parsed

### Requirement: Collector lifecycle and shutdown

The collector SHALL stop the target and close the snapshot channel when
it is stopped, when its context is cancelled, or when the target exits.
The temporary build artifact SHALL be removed on shutdown.

#### Scenario: Stop terminates the target
- **WHEN** the collector is stopped or its context is cancelled
- **THEN** the target process is killed, the temporary binary is removed, and the snapshot channel is closed

#### Scenario: Target exit ends the stream
- **WHEN** the target process exits on its own
- **THEN** the collector closes the snapshot channel after the final line is read

### Requirement: Maintain current state and rolling history

The monitor state SHALL store the latest snapshot and a rolling history
bounded to the most recent 60 snapshots. Reads SHALL return copies so
callers cannot mutate internal state, and concurrent updates and reads
SHALL be safe.

#### Scenario: History bounded to the most recent points
- **WHEN** more than 60 snapshots have been recorded
- **THEN** only the most recent 60 are retained in history

#### Scenario: Reads return the latest plus a history copy
- **WHEN** the current state is read
- **THEN** it returns the latest snapshot together with an independent copy of the retained history
