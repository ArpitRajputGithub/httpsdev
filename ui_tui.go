package main

// runTUI is implemented in Task 11.
func runTUI(listenAddr, upstream string, events <-chan Event) {
	// Fallback for now: just drain the channel silently.
	for range events {
	}
	_ = listenAddr
	_ = upstream
}
