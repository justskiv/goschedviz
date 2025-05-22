// Command monitor provides real-time visualization of Go scheduler metrics.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/JustSkiv/goschedviz/internal/collector/godebug"
	"github.com/JustSkiv/goschedviz/internal/domain"
	"github.com/JustSkiv/goschedviz/internal/ui"
	"github.com/JustSkiv/goschedviz/internal/ui/termui"
)

type collector interface {
	Start(ctx context.Context) (<-chan domain.SchedulerSnapshot, error)
	Stop() error
}

type presenter interface {
	Start() error
	Stop()
	Update(data ui.UIData)
	Done() <-chan struct{}
}

func main() {
	var (
		targetPath = flag.String("target", "", "Path to Go program to monitor")
		period     = flag.Int("period", 1000, "GODEBUG schedtrace period in milliseconds")
		showHelp   = flag.Bool("help", false, "Prints usage instructions")
	)
	flag.BoolVar(showHelp, "h", false, "Prints usage instructions (alias for -help)")

	flag.Parse()

	if *showHelp {
		flag.Usage()
		os.Exit(0)
	}

	if *targetPath == "" {
		fmt.Println("Please specify target program path with -target flag")
		flag.Usage()
		os.Exit(1)
	}

	// Create collector
	collector := godebug.New(*targetPath, *period)

	// Create UI
	presenter := termui.New()
	if err := presenter.Start(); err != nil {
		log.Fatalf("Failed to initialize UI: %v", err)
	}
	defer presenter.Stop()

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle Ctrl+C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	go func() {
		<-sigChan
		cancel()
	}()

	if err := monitorScheduler(ctx, collector, presenter); err != nil {
		log.Println("Error:", err)
		return
	}

}

func monitorScheduler(ctx context.Context, c collector, p presenter) error {
	snapshots, err := c.Start(ctx)
	if err != nil {
		return fmt.Errorf("failed to start collector: %w", err)
	}
	defer func() {
		if err := c.Stop(); err != nil {
			log.Println("Failed to stop collector:", err)
		}
	}()

	state := &domain.MonitorState{}
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case snapshot, ok := <-snapshots:
			if !ok {
				return nil
			}
			state.Update(snapshot)

		case <-ticker.C:
			latest, history := state.GetSnapshot()
			uiData := convertToUIData(latest, history)
			p.Update(uiData)

		case <-p.Done():
			return nil

		case <-ctx.Done():
			return nil
		}
	}
}

// convertToUIData converts domain data to UI-specific format
func convertToUIData(latest domain.SchedulerSnapshot, history []domain.SchedulerSnapshot) ui.UIData {
	// Calculate max values for gauges
	maxGRQ, maxGoroutines, maxThreads, maxIdleProcs := 0, 0, 0, 0
	histValues := make([]ui.HistoricalValues, len(history))

	for i, h := range history {
		if h.RunQueue > maxGRQ {
			maxGRQ = h.RunQueue
		}
		if h.Goroutines > maxGoroutines {
			maxGoroutines = h.Goroutines
		}
		if h.Threads > maxThreads {
			maxThreads = h.Threads
		}
		if h.IdleProcs > maxIdleProcs {
			maxIdleProcs = h.IdleProcs
		}

		histValues[i] = ui.HistoricalValues{
			TimeMs:     h.TimeMs,
			GRQ:        h.RunQueue,
			LRQSum:     h.LRQSum,
			IdleProcs:  h.IdleProcs,
			Threads:    h.Threads,
			Goroutines: h.Goroutines,
		}
	}

	// Ensure non-zero max values for gauges
	if maxGRQ == 0 {
		maxGRQ = 1
	}
	if maxGoroutines == 0 {
		maxGoroutines = 1
	}
	if maxThreads == 0 {
		maxThreads = 1
	}
	if maxIdleProcs == 0 {
		maxIdleProcs = 1
	}

	result := ui.UIData{
		Current: ui.CurrentValues{
			TimeMs:          latest.TimeMs,
			GoMaxProcs:      latest.GoMaxProcs,
			IdleProcs:       latest.IdleProcs,
			Threads:         latest.Threads,
			SpinningThreads: latest.SpinningThreads,
			NeedSpinning:    latest.NeedSpinning,
			IdleThreads:     latest.IdleThreads,
			RunQueue:        latest.RunQueue,
			LRQSum:          latest.LRQSum,
			NumP:            len(latest.LRQ),
			LRQ:             latest.LRQ,
			Goroutines:      latest.Goroutines,
		},
		Gauges: ui.GaugeValues{
			GRQ: struct {
				Current int
				Max     int
			}{
				Current: latest.RunQueue,
				Max:     maxGRQ,
			},
			Goroutines: struct {
				Current int
				Max     int
			}{
				Current: latest.Goroutines,
				Max:     maxGoroutines,
			},
			Threads: struct {
				Current int
				Max     int
			}{
				Current: latest.Threads,
				Max:     maxThreads,
			},
			IdleProcs: struct {
				Current int
				Max     int
			}{
				Current: latest.IdleProcs,
				Max:     maxIdleProcs,
			},
		},
	}

	result.History.Raw = histValues

	return result
}
