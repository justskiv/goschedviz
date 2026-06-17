## ADDED Requirements

### Requirement: Command-line flags

The CLI SHALL accept a `-target` flag with the path to the Go program to
monitor and a `-period` flag for the schedtrace period in milliseconds
(default 1000). A target SHALL be required.

#### Scenario: Missing target stops with a message
- **WHEN** the program is run without a `-target` value
- **THEN** it prints a message asking for the target path and exits with a non-zero status

#### Scenario: Default trace period
- **WHEN** the program is run without a `-period` value
- **THEN** the schedtrace period defaults to 1000 milliseconds

### Requirement: Wire components and run the refresh loop

The CLI SHALL connect the collector, monitor state, and presenter:
incoming snapshots update the state, and on a fixed refresh tick the
latest state and history are converted to UI data and pushed to the
presenter.

#### Scenario: Snapshots update state and the UI refreshes on tick
- **WHEN** the collector delivers snapshots and a refresh tick fires
- **THEN** the state is updated from snapshots and the presenter is updated with the latest values and history

### Requirement: Graceful shutdown

The CLI SHALL shut down cleanly when it receives an interrupt signal,
when the user quits the UI, or when the target exits, stopping the
collector and closing the UI.

#### Scenario: Interrupt signal shuts down
- **WHEN** the process receives an interrupt signal
- **THEN** the context is cancelled, the collector is stopped, and the UI is closed

#### Scenario: UI quit shuts down
- **WHEN** the presenter signals done because the user quit
- **THEN** the refresh loop exits and the collector is stopped

#### Scenario: Target exit shuts down
- **WHEN** the snapshot channel closes because the target exited
- **THEN** the refresh loop exits and the collector is stopped
