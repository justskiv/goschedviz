package main

import (
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/JustSkiv/goschedviz/internal/domain"
	"github.com/JustSkiv/goschedviz/internal/ui"
)

func TestConvertToUIData_Simple(t *testing.T) {
	// Prepare test data
	latest := domain.SchedulerSnapshot{
		TimeMs:     1000,
		GoMaxProcs: 2,
		LRQ:        []int{1, 2},
		LRQSum:     3,
		Goroutines: 100,
	}

	history := []domain.SchedulerSnapshot{latest}

	// Call the function under test
	result := convertToUIData(latest, history)

	// Assert results
	assert.Equal(t, 1000, result.Current.TimeMs)
	assert.Equal(t, 2, result.Current.GoMaxProcs)
	assert.Equal(t, []int{1, 2}, result.Current.LRQ)
	assert.Equal(t, 3, result.Current.LRQSum)
	assert.Equal(t, 100, result.Current.Goroutines)
}

func TestConvertToUIData2(t *testing.T) {
	tests := []struct {
		name     string
		latest   domain.SchedulerSnapshot
		history  []domain.SchedulerSnapshot
		expected struct {
			maxGRQ        int
			maxGoroutines int
			maxThreads    int
			maxIdleProcs  int
		}
	}{
		{
			name: "empty_history",
			latest: domain.SchedulerSnapshot{
				TimeMs:     1000,
				GoMaxProcs: 4,
				IdleProcs:  2,
				Threads:    8,
				RunQueue:   5,
				LRQ:        []int{1, 2, 1, 0},
				LRQSum:     4,
				Goroutines: 100,
			},
			history: nil,
			expected: struct {
				maxGRQ        int
				maxGoroutines int
				maxThreads    int
				maxIdleProcs  int
			}{
				maxGRQ:        1,
				maxGoroutines: 1,
				maxThreads:    1,
				maxIdleProcs:  1,
			},
		},
		{
			name: "single_processor",
			latest: domain.SchedulerSnapshot{
				TimeMs:     1000,
				GoMaxProcs: 1,
				IdleProcs:  0,
				Threads:    2,
				RunQueue:   5,
				LRQ:        []int{3},
				LRQSum:     3,
				Goroutines: 150,
			},
			history: []domain.SchedulerSnapshot{
				{TimeMs: 0, RunQueue: 2, Threads: 1, LRQ: []int{1}, LRQSum: 1, Goroutines: 100},
				{TimeMs: 500, RunQueue: 3, Threads: 2, LRQ: []int{2}, LRQSum: 2, Goroutines: 120},
				{TimeMs: 1000, RunQueue: 5, Threads: 2, LRQ: []int{3}, LRQSum: 3, Goroutines: 150},
			},
			expected: struct {
				maxGRQ        int
				maxGoroutines int
				maxThreads    int
				maxIdleProcs  int
			}{
				maxGRQ:        5,
				maxGoroutines: 150,
				maxThreads:    2,
				maxIdleProcs:  1,
			},
		},
		{
			name: "growing_load",
			latest: domain.SchedulerSnapshot{
				TimeMs:     3000,
				GoMaxProcs: 4,
				IdleProcs:  0,
				Threads:    16,
				RunQueue:   15,
				LRQ:        []int{5, 5, 5, 5},
				LRQSum:     20,
				Goroutines: 500,
			},
			history: []domain.SchedulerSnapshot{
				{TimeMs: 1000, RunQueue: 5, Threads: 8, IdleProcs: 2, LRQSum: 10, Goroutines: 200},
				{TimeMs: 2000, RunQueue: 10, Threads: 12, IdleProcs: 1, LRQSum: 15, Goroutines: 350},
				{TimeMs: 3000, RunQueue: 15, Threads: 16, IdleProcs: 0, LRQSum: 20, Goroutines: 500},
			},
			expected: struct {
				maxGRQ        int
				maxGoroutines int
				maxThreads    int
				maxIdleProcs  int
			}{
				maxGRQ:        15,
				maxGoroutines: 500,
				maxThreads:    16,
				maxIdleProcs:  2,
			},
		},
		{
			name: "zero_values",
			latest: domain.SchedulerSnapshot{
				TimeMs:     1000,
				GoMaxProcs: 2,
				IdleProcs:  0,
				Threads:    0,
				RunQueue:   0,
				LRQ:        []int{0, 0},
				LRQSum:     0,
				Goroutines: 0,
			},
			history: []domain.SchedulerSnapshot{
				{TimeMs: 1000, RunQueue: 0, Threads: 0, LRQSum: 0, Goroutines: 0},
			},
			expected: struct {
				maxGRQ        int
				maxGoroutines int
				maxThreads    int
				maxIdleProcs  int
			}{
				maxGRQ:        1,
				maxGoroutines: 1,
				maxThreads:    1,
				maxIdleProcs:  1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertToUIData(tt.latest, tt.history)

			// Verify current values
			assert.Equal(t, tt.latest.TimeMs, result.Current.TimeMs)
			assert.Equal(t, tt.latest.GoMaxProcs, result.Current.GoMaxProcs)
			assert.Equal(t, tt.latest.IdleProcs, result.Current.IdleProcs)
			assert.Equal(t, tt.latest.Threads, result.Current.Threads)
			assert.Equal(t, tt.latest.RunQueue, result.Current.RunQueue)
			assert.Equal(t, tt.latest.LRQ, result.Current.LRQ)
			assert.Equal(t, tt.latest.LRQSum, result.Current.LRQSum)
			assert.Equal(t, tt.latest.Goroutines, result.Current.Goroutines)
			assert.Equal(t, len(tt.latest.LRQ), result.Current.NumP)

			// Verify max values
			assert.Equal(t, tt.expected.maxGRQ, result.Gauges.GRQ.Max)
			assert.Equal(t, tt.expected.maxGoroutines, result.Gauges.Goroutines.Max)
			assert.Equal(t, tt.expected.maxThreads, result.Gauges.Threads.Max)
			assert.Equal(t, tt.expected.maxIdleProcs, result.Gauges.IdleProcs.Max)

			// Verify history conversion
			if tt.history != nil {
				assert.Equal(t, len(tt.history), len(result.History.Raw))
				for i, h := range tt.history {
					assert.Equal(t, h.TimeMs, result.History.Raw[i].TimeMs)
					assert.Equal(t, h.RunQueue, result.History.Raw[i].GRQ)
					assert.Equal(t, h.LRQSum, result.History.Raw[i].LRQSum)
					assert.Equal(t, h.IdleProcs, result.History.Raw[i].IdleProcs)
					assert.Equal(t, h.Threads, result.History.Raw[i].Threads)
					assert.Equal(t, h.Goroutines, result.History.Raw[i].Goroutines)
				}
			} else {
				assert.Empty(t, result.History.Raw)
			}
		})
	}
}

func TestMainHelpFlag(t *testing.T) {
	// Build the goschedviz binary
	cmdBuild := exec.Command("go", "build", "-o", "goschedviz_test_binary", ".")
	// Explicitly set the directory for the build command if main_test.go is not in the main package's directory
	// For example, if main.go is in cmd/goschedviz and main_test.go is also there,
	// the current directory "." is fine. If main_test.go is in a subdirectory, adjust accordingly.
	// cmdBuild.Dir = ".." // Or the correct path to the main package
	err := cmdBuild.Run()
	if err != nil {
		t.Fatalf("Failed to build goschedviz binary: %v", err)
	}
	defer os.Remove("goschedviz_test_binary") // Clean up the binary after the test

	// Execute the compiled binary with the --help argument
	cmdRun := exec.Command("./goschedviz_test_binary", "--help")
	output, err := cmdRun.CombinedOutput() // CombinedOutput captures both stdout and stderr

	// Check if the program exited successfully (status code 0)
	// For --help, we expect a successful exit.
	if exitErr, ok := err.(*exec.ExitError); ok {
		// The program exited with an error code
		t.Fatalf("Expected exit code 0, but got %d. Output:\n%s", exitErr.ExitCode(), string(output))
	} else if err != nil {
		// Another error occurred (e.g., binary not found, though build should prevent this)
		t.Fatalf("Error running goschedviz with --help: %v. Output:\n%s", err, string(output))
	}

	// Verify that the output to stdout contains the usage string
	expectedOutputSubstring := "Usage of ./goschedviz_test_binary:"
	if !strings.Contains(string(output), expectedOutputSubstring) {
		t.Errorf("Expected output to contain '%s', but got:\n%s", expectedOutputSubstring, string(output))
	}

	// Test with -h alias
	cmdRunAlias := exec.Command("./goschedviz_test_binary", "-h")
	outputAlias, errAlias := cmdRunAlias.CombinedOutput()

	if exitErr, ok := errAlias.(*exec.ExitError); ok {
		t.Fatalf("Expected exit code 0 for -h, but got %d. Output:\n%s", exitErr.ExitCode(), string(outputAlias))
	} else if errAlias != nil {
		t.Fatalf("Error running goschedviz with -h: %v. Output:\n%s", errAlias, string(outputAlias))
	}

	if !strings.Contains(string(outputAlias), expectedOutputSubstring) {
		t.Errorf("Expected output for -h to contain '%s', but got:\n%s", expectedOutputSubstring, string(outputAlias))
	}
}

func TestConvertToUIData_Complex(t *testing.T) {
	tests := []struct {
		name    string
		latest  domain.SchedulerSnapshot
		history []domain.SchedulerSnapshot
		want    ui.UIData
	}{
		{
			name: "empty_history_non_zero_latest",
			latest: domain.SchedulerSnapshot{
				TimeMs:          1000,
				GoMaxProcs:      4,
				IdleProcs:       2,
				Threads:         8,
				SpinningThreads: 1,
				NeedSpinning:    0,
				IdleThreads:     3,
				RunQueue:        5,
				LRQ:             []int{1, 2, 1, 0},
				LRQSum:          4,
				Goroutines:      100,
			},
			history: nil,
			want: ui.UIData{
				Current: ui.CurrentValues{
					TimeMs:          1000,
					GoMaxProcs:      4,
					IdleProcs:       2,
					Threads:         8,
					SpinningThreads: 1,
					NeedSpinning:    0,
					IdleThreads:     3,
					RunQueue:        5,
					LRQSum:          4,
					NumP:            4,
					LRQ:             []int{1, 2, 1, 0},
					Goroutines:      100,
				},
				Gauges: ui.GaugeValues{
					GRQ: struct{ Current, Max int }{
						Current: 5,
						Max:     1,
					},
					Goroutines: struct{ Current, Max int }{
						Current: 100,
						Max:     1,
					},
					Threads: struct{ Current, Max int }{
						Current: 8,
						Max:     1,
					},
					IdleProcs: struct{ Current, Max int }{
						Current: 2,
						Max:     1,
					},
				},
			},
		},
		{
			name: "increasing_load_pattern",
			latest: domain.SchedulerSnapshot{
				TimeMs:     3000,
				GoMaxProcs: 8,
				IdleProcs:  1,
				Threads:    16,
				RunQueue:   20,
				LRQ:        []int{4, 4, 4, 4, 4, 4, 4, 4},
				LRQSum:     32,
				Goroutines: 500,
			},
			history: []domain.SchedulerSnapshot{
				{
					TimeMs:     1000,
					GoMaxProcs: 8,
					IdleProcs:  6,
					Threads:    10,
					RunQueue:   5,
					LRQSum:     8,
					Goroutines: 100,
				},
				{
					TimeMs:     2000,
					GoMaxProcs: 8,
					IdleProcs:  3,
					Threads:    12,
					RunQueue:   10,
					LRQSum:     16,
					Goroutines: 250,
				},
				{
					TimeMs:     3000,
					GoMaxProcs: 8,
					IdleProcs:  1,
					Threads:    16,
					RunQueue:   20,
					LRQSum:     32,
					Goroutines: 500,
				},
			},
			want: ui.UIData{
				Current: ui.CurrentValues{
					TimeMs:     3000,
					GoMaxProcs: 8,
					IdleProcs:  1,
					Threads:    16,
					RunQueue:   20,
					NumP:       8,
					LRQ:        []int{4, 4, 4, 4, 4, 4, 4, 4},
					LRQSum:     32,
					Goroutines: 500,
				},
				History: struct {
					Raw    []ui.HistoricalValues
					Scaled []ui.HistoricalValues
				}{
					Raw: []ui.HistoricalValues{
						{TimeMs: 1000, GRQ: 5, LRQSum: 8, IdleProcs: 6, Threads: 10, Goroutines: 100},
						{TimeMs: 2000, GRQ: 10, LRQSum: 16, IdleProcs: 3, Threads: 12, Goroutines: 250},
						{TimeMs: 3000, GRQ: 20, LRQSum: 32, IdleProcs: 1, Threads: 16, Goroutines: 500},
					},
				},
				Gauges: ui.GaugeValues{
					GRQ: struct{ Current, Max int }{
						Current: 20,
						Max:     20,
					},
					Goroutines: struct{ Current, Max int }{
						Current: 500,
						Max:     500,
					},
					Threads: struct{ Current, Max int }{
						Current: 16,
						Max:     16,
					},
					IdleProcs: struct{ Current, Max int }{
						Current: 1,
						Max:     6,
					},
				},
			},
		},
		{
			name: "all_zero_metrics",
			latest: domain.SchedulerSnapshot{
				TimeMs:     0,
				GoMaxProcs: 1,
				LRQ:        []int{0},
			},
			history: []domain.SchedulerSnapshot{
				{TimeMs: 0, GoMaxProcs: 1, LRQ: []int{0}},
			},
			want: ui.UIData{
				Current: ui.CurrentValues{
					GoMaxProcs: 1,
					NumP:       1,
					LRQ:        []int{0},
				},
				Gauges: ui.GaugeValues{
					GRQ:        struct{ Current, Max int }{Current: 0, Max: 1},
					Goroutines: struct{ Current, Max int }{Current: 0, Max: 1},
					Threads:    struct{ Current, Max int }{Current: 0, Max: 1},
					IdleProcs:  struct{ Current, Max int }{Current: 0, Max: 1},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertToUIData(tt.latest, tt.history)

			// Check Current values
			assert.Equal(t, tt.want.Current, got.Current, "Current values mismatch")

			// Check History values if present
			if tt.want.History.Raw != nil {
				assert.Equal(t, tt.want.History.Raw, got.History.Raw, "History values mismatch")
			}

			// Check Gauge values
			assert.Equal(t, tt.want.Gauges, got.Gauges, "Gauge values mismatch")
		})
	}
}
