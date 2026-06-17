## ADDED Requirements

### Requirement: Presenter lifecycle

The dashboard SHALL implement a presenter that can be started, updated
with new data, stopped, and observed for termination via a done channel.
Starting SHALL initialize the terminal and all widgets; if the terminal
cannot be initialized, starting SHALL return an error.

#### Scenario: Start initializes the UI
- **WHEN** the presenter is started successfully
- **THEN** the terminal and all widgets are initialized and event handling begins

#### Scenario: Terminal initialization failure is reported
- **WHEN** the underlying terminal fails to initialize
- **THEN** Start returns an error and the UI does not run

### Requirement: Render current scheduler values

On each update the dashboard SHALL display the current scheduler values
— GoMaxProcs, idle processors, threads, spinning threads, need-spinning,
idle threads, global run queue, local run queue sum, number of P, and
goroutine count — in the values table and info box.

#### Scenario: Update renders current values
- **WHEN** the presenter receives new UI data
- **THEN** the table and info box reflect that snapshot's current values

### Requirement: Visualize per-P local run queues

The dashboard SHALL display the per-processor local run queue lengths as
a bar chart, one bar per P.

#### Scenario: Bar chart reflects local run queues
- **WHEN** the presenter receives UI data with per-P local run queue lengths
- **THEN** the bar chart shows a bar per P sized to each queue length

### Requirement: Gauges for key metrics

The dashboard SHALL display gauges for the global run queue, goroutine
count, thread count, and idle processors, each showing the current value
relative to its observed maximum.

#### Scenario: Gauges show current against maximum
- **WHEN** the presenter receives UI data with gauge current and max values
- **THEN** each gauge renders its current value scaled against its maximum

### Requirement: History plots

The dashboard SHALL plot historical metrics on both a linear and a
logarithmic chart and SHALL show a legend identifying the plotted
series.

#### Scenario: Plots updated from history
- **WHEN** the presenter receives UI data containing historical values
- **THEN** the linear and logarithmic plots are updated and the legend identifies the series

### Requirement: Quit on key or interrupt

The dashboard SHALL signal termination through its done channel when the
user presses `q` or `Ctrl+C`.

#### Scenario: Quit key terminates the UI
- **WHEN** the user presses `q` or `Ctrl+C`
- **THEN** the done channel is closed to signal that the UI should shut down

### Requirement: Handle terminal resize

The dashboard SHALL adapt its layout to the terminal size when a resize
event occurs.

#### Scenario: Resize re-lays out the grid
- **WHEN** a terminal resize event is received
- **THEN** the layout is resized to the new dimensions and re-rendered
