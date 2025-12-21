package status

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// ANSI color codes
const (
	Reset        = "\033[0m"
	Green        = "\033[32m"
	Yellow       = "\033[33m"
	Blue         = "\033[34m"
	Red          = "\033[31m"
	Gray         = "\033[90m"
	Bold         = "\033[1m"
	ClearLine    = "\033[2K"
	CursorUp     = "\033[1A"
	SaveCursor   = "\033[s"
	RestoreCursor = "\033[u"
)

// State represents the current state of a workflow step
type State int

const (
	StatePending State = iota
	StateRunning
	StateComplete
	StateFailed
)

// Step represents a single step in the workflow
type Step struct {
	Name        string
	Description string
	State       State
	StartTime   time.Time
	EndTime     time.Time
}

// Reporter provides inline status reporting for CLI operations
type Reporter struct {
	steps      []Step
	currentIdx int
	writer     io.Writer
	mu         sync.Mutex
	colored    bool
	lastOutput string
}

// NewReporter creates a new status reporter
func NewReporter(colored bool) *Reporter {
	return &Reporter{
		steps:   make([]Step, 0),
		writer:  os.Stdout,
		colored: colored,
	}
}

// AddStep adds a new step to the workflow
func (r *Reporter) AddStep(name, description string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.steps = append(r.steps, Step{
		Name:        name,
		Description: description,
		State:       StatePending,
	})
}

// StartStep marks a step as running
func (r *Reporter) StartStep(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	for i := range r.steps {
		if r.steps[i].Name == name {
			r.steps[i].State = StateRunning
			r.steps[i].StartTime = time.Now()
			r.currentIdx = i
			r.render()
			return
		}
	}
}

// CompleteStep marks a step as complete
func (r *Reporter) CompleteStep(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	for i := range r.steps {
		if r.steps[i].Name == name {
			r.steps[i].State = StateComplete
			r.steps[i].EndTime = time.Now()
			r.render()
			return
		}
	}
}

// FailStep marks a step as failed
func (r *Reporter) FailStep(name string, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	for i := range r.steps {
		if r.steps[i].Name == name {
			r.steps[i].State = StateFailed
			r.steps[i].EndTime = time.Now()
			r.render()
			if err != nil {
				fmt.Fprintf(r.writer, "\n%sError: %v%s\n", r.color(Red), err, r.color(Reset))
			}
			return
		}
	}
}

// render updates the display with current status
func (r *Reporter) render() {
	// Clear previous output if in interactive mode
	if r.lastOutput != "" && r.isTerminal() {
		// Move cursor up and clear lines
		lineCount := len(r.steps) + 1
		for i := 0; i < lineCount; i++ {
			fmt.Fprintf(r.writer, "%s%s\r", CursorUp, ClearLine)
		}
	}
	
	output := r.buildOutput()
	fmt.Fprint(r.writer, output)
	r.lastOutput = output
}

// buildOutput creates the status display
func (r *Reporter) buildOutput() string {
	var output string
	
	// Title
	output += fmt.Sprintf("\n%s%sIDPBuilder Progress%s\n", r.color(Bold), r.color(Blue), r.color(Reset))
	
	// Steps
	for i, step := range r.steps {
		symbol := r.getSymbol(step.State)
		color := r.getColor(step.State)
		
		status := ""
		if step.State == StateRunning {
			status = fmt.Sprintf(" %s(in progress)%s", r.color(Gray), r.color(Reset))
		} else if step.State == StateComplete && !step.EndTime.IsZero() && !step.StartTime.IsZero() {
			duration := step.EndTime.Sub(step.StartTime).Round(time.Millisecond)
			status = fmt.Sprintf(" %s(%s)%s", r.color(Gray), duration, r.color(Reset))
		}
		
		// Format: [✓] Step description (status)
		output += fmt.Sprintf("  %s%s%s %s%s\n", 
			r.color(color), 
			symbol, 
			r.color(Reset), 
			step.Description,
			status)
		
		// Add separator after current running step
		if i == r.currentIdx && step.State == StateRunning {
			output += fmt.Sprintf("  %s│%s\n", r.color(Blue), r.color(Reset))
		}
	}
	
	return output
}

// getSymbol returns the symbol for a state
func (r *Reporter) getSymbol(state State) string {
	switch state {
	case StatePending:
		return "○"
	case StateRunning:
		return "●"
	case StateComplete:
		return "✓"
	case StateFailed:
		return "✗"
	default:
		return "○"
	}
}

// getColor returns the color for a state
func (r *Reporter) getColor(state State) string {
	switch state {
	case StatePending:
		return Gray
	case StateRunning:
		return Blue
	case StateComplete:
		return Green
	case StateFailed:
		return Red
	default:
		return Reset
	}
}

// color returns the ANSI color code if colored output is enabled
func (r *Reporter) color(code string) string {
	if r.colored {
		return code
	}
	return ""
}

// isTerminal checks if output is a terminal
func (r *Reporter) isTerminal() bool {
	if f, ok := r.writer.(*os.File); ok {
		fileInfo, err := f.Stat()
		if err != nil {
			return false
		}
		return (fileInfo.Mode() & os.ModeCharDevice) != 0
	}
	return false
}

// Summary prints a final summary
func (r *Reporter) Summary() {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	// Ensure we have a clean final render
	r.render()
	
	// Count states
	completed := 0
	failed := 0
	for _, step := range r.steps {
		if step.State == StateComplete {
			completed++
		} else if step.State == StateFailed {
			failed++
		}
	}
	
	if failed > 0 {
		fmt.Fprintf(r.writer, "\n%s%s✗ Build failed: %d/%d steps completed%s\n", 
			r.color(Bold), r.color(Red), completed, len(r.steps), r.color(Reset))
	} else {
		fmt.Fprintf(r.writer, "\n%s%s✓ Build completed successfully: %d/%d steps%s\n", 
			r.color(Bold), r.color(Green), completed, len(r.steps), r.color(Reset))
	}
}
