## 1. Verify scheduler-monitoring against code

- [x] 1.1 Confirm collector config validation (period > 0; path non-empty, existing, file, `.go`) matches `internal/collector/godebug/collector.go`
- [x] 1.2 Confirm the target is built and run with `GODEBUG=schedtrace=<period>` and stdin attached
- [x] 1.3 Confirm parser fields and per-P LRQ extraction match the `SCHED` regex in `parser.go`
- [x] 1.4 Confirm `PROCMETR num_goroutines=<n>` parsing and carry-over into snapshots
- [x] 1.5 Confirm snapshot validation rules (GoMaxProcs, idle/LRQ/thread consistency, all-zero reject)
- [x] 1.6 Confirm channel emission and shutdown (Stop/ctx cancel/target exit close the channel; temp binary removed)
- [x] 1.7 Confirm `MonitorState` caps history at 60 and returns copies (`internal/domain/scheduler.go`)

## 2. Verify terminal-dashboard against code

- [x] 2.1 Confirm `Presenter` lifecycle (Start/Stop/Update/Done) and start-failure error in `internal/ui/termui/renderer.go`
- [x] 2.2 Confirm table/info, LRQ bar chart, gauges, linear+log plots, and legend update on `Update`
- [x] 2.3 Confirm `q`/`Ctrl+C` close Done and resize re-lays out the grid

## 3. Verify goroutine-metrics-reporter against code

- [x] 3.1 Confirm periodic stderr output format `PROCMETR num_goroutines=<n>` in `pkg/metrics/metrics.go`
- [x] 3.2 Confirm background lifecycle and idempotent Start/Stop (`sync.Once`)

## 4. Verify cli against code

- [x] 4.1 Confirm `-target` (required) and `-period` (default 1000) flags in `cmd/goschedviz/main.go`
- [x] 4.2 Confirm wiring collector → state → presenter and the 500ms refresh tick
- [x] 4.3 Confirm graceful shutdown on interrupt signal, UI quit, and target exit

## 5. Validate and finalize

- [x] 5.1 Run `make test` and confirm the suite passes (specs reflect tested behavior)
- [x] 5.2 Run `openspec validate add-baseline-specs --strict` and resolve any issues
- [x] 5.3 Archive the change to populate `openspec/specs/` with the four capabilities
