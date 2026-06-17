## ADDED Requirements

### Requirement: Periodic goroutine reporting

The reporter SHALL, once started, emit the live goroutine count to
standard error at a fixed interval using a line prefixed with `PROCMETR`
in the form `PROCMETR num_goroutines=<n>`, so the monitoring tool can
parse it.

#### Scenario: Reports at the configured interval
- **WHEN** a reporter created with an interval is started
- **THEN** it writes the current goroutine count to stderr once per interval

#### Scenario: Output uses the parseable PROCMETR format
- **WHEN** the reporter emits a metrics line
- **THEN** the line is `PROCMETR num_goroutines=<n>` where `<n>` is the current goroutine count

### Requirement: Safe lifecycle

The reporter SHALL run in a background goroutine until stopped, and
SHALL tolerate repeated start and stop calls so that only the first of
each takes effect.

#### Scenario: Stop halts reporting
- **WHEN** the reporter is stopped
- **THEN** it stops emitting metrics lines

#### Scenario: Repeated stop is safe
- **WHEN** Stop is called more than once
- **THEN** only the first call takes effect and no panic occurs
