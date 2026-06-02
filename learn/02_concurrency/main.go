// learn/02_concurrency/main.go
package main

import (
	"context"
	"fmt"
	"time"
)

// ProgressEvent mirrors the real app's type.
type ProgressEvent struct {
	Type      string // "searching" | "processing" | "done"
	Processed int
	Total     int
}

// simulateSearch mimics a Jira worklog fetch.
func simulateSearch(ctx context.Context, total int, progress chan<- ProgressEvent) {
	defer close(progress) // ALWAYS close when done — reader loop ends on close

	for i := 1; i <= total; i++ {
		select {
		case <-ctx.Done():
			fmt.Println("worker: cancelled at item", i)
			return
		default:
		}

		time.Sleep(100 * time.Millisecond)

		progress <- ProgressEvent{
			Type:      "processing",
			Processed: i,
			Total:     total,
		}
	}
	progress <- ProgressEvent{Type: "done"}
}

func main() {
	fmt.Println("--- Example 1: Normal completion ---")
	runSearch(context.Background(), 5)

	fmt.Println("\n--- Example 2: Cancellation after 2 items ---")
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(250 * time.Millisecond) // cancel after ~2 items
		cancel()
	}()
	runSearch(ctx, 10)
}

func runSearch(ctx context.Context, total int) {
	progress := make(chan ProgressEvent, 10) // buffered — worker never blocks

	go simulateSearch(ctx, total, progress)

	for ev := range progress {
		switch ev.Type {
		case "processing":
			fmt.Printf("  processed %d / %d\n", ev.Processed, ev.Total)
		case "done":
			fmt.Println("  done!")
		}
	}
	fmt.Println("search complete")
}
