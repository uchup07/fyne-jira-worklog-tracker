// learn/01_structs/main.go
package main

import (
	"fmt"
	"time"
)

// --- Structs and methods ---

// WorklogItem represents a single worklog entry (mirrors the real app type).
type WorklogItem struct {
	IssueKey         string
	Author           string
	TimeSpentSeconds int
	Started          time.Time
}

// Hours returns the time spent as fractional hours.
func (w WorklogItem) Hours() float64 {
	return float64(w.TimeSpentSeconds) / 3600
}

// WorklogGroup groups items by work reference (mirrors the real app type).
type WorklogGroup struct {
	WorkReference string
	Items         []WorklogItem
}

// TotalHours sums hours across all items in the group.
func (g WorklogGroup) TotalHours() float64 {
	total := 0.0
	for _, item := range g.Items {
		total += item.Hours()
	}
	return total
}

// --- Interfaces ---

// Summariser is any type that can produce a one-line summary string.
type Summariser interface {
	Summary() string
}

func (g WorklogGroup) Summary() string {
	return fmt.Sprintf("%s: %.1fh (%d entries)", g.WorkReference, g.TotalHours(), len(g.Items))
}

// printAll prints the summary of anything that implements Summariser.
func printAll(items []Summariser) {
	for _, item := range items {
		fmt.Println(item.Summary())
	}
}

// --- Error handling ---

// parseHours converts a "Xh Ym" string to total seconds.
func parseHours(s string) (int, error) {
	var h, m int
	_, err := fmt.Sscanf(s, "%dh %dm", &h, &m)
	if err != nil {
		return 0, fmt.Errorf("parseHours: invalid format %q (expected e.g. '2h 30m'): %w", s, err)
	}
	return h*3600 + m*60, nil
}

func main() {
	// Build two groups
	groups := []WorklogGroup{
		{
			WorkReference: "CR-001",
			Items: []WorklogItem{
				{IssueKey: "JSW-1", Author: "Alice", TimeSpentSeconds: 7200, Started: time.Now()},
				{IssueKey: "JSW-2", Author: "Bob", TimeSpentSeconds: 3600, Started: time.Now()},
			},
		},
		{
			WorkReference: "CR-002",
			Items: []WorklogItem{
				{IssueKey: "JSW-3", Author: "Alice", TimeSpentSeconds: 5400, Started: time.Now()},
			},
		},
	}

	// Use interface — WorklogGroup satisfies Summariser
	summaries := make([]Summariser, len(groups))
	for i, g := range groups {
		summaries[i] = g
	}
	printAll(summaries)

	// Error handling
	secs, err := parseHours("2h 30m")
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Printf("2h 30m = %d seconds\n", secs)
	}

	_, err = parseHours("bad input")
	if err != nil {
		fmt.Println("Expected error:", err)
	}
}
