package main

import (
	"fmt"
	"io"
	"time"

	"github.com/fatih/color"
)

// runLogUI prints the startup banner then one line per incoming Event.
// Returns after events channel is closed. Prints a summary on exit.
func runLogUI(w io.Writer, listenAddr, upstream string, events <-chan Event) {
	color.NoColor = !isTerminal(w)

	fmt.Fprintln(w)
	fmt.Fprintf(w, "  %s httpsdev v%s\n\n", color.CyanString("▲"), Version)
	fmt.Fprintf(w, "  %s Local:    %s\n", color.GreenString("➜"), color.New(color.Bold).Sprintf("https://%s", listenAddr))
	fmt.Fprintf(w, "  %s Upstream: %s\n", color.GreenString("➜"), upstream)
	fmt.Fprintf(w, "  %s Cert:     mkcert · valid 30 days\n\n", color.GreenString("➜"))
	fmt.Fprintf(w, "  press Ctrl+C to quit\n\n")

	var total, errors int
	var totalDur time.Duration
	start := time.Now()

	for ev := range events {
		total++
		totalDur += ev.Duration
		if ev.Status >= 400 {
			errors++
		}
		fmt.Fprintln(w, formatEvent(ev))
	}

	if total == 0 {
		return
	}
	avg := totalDur / time.Duration(total)
	fmt.Fprintf(w, "\nserved %d requests · %d errors · avg %s · uptime %s\n",
		total, errors, avg.Round(time.Millisecond), time.Since(start).Round(time.Second))
}

func formatEvent(ev Event) string {
	statusColor := color.New(color.FgGreen)
	switch {
	case ev.Status >= 500:
		statusColor = color.New(color.FgRed, color.Bold)
	case ev.Status >= 400:
		statusColor = color.New(color.FgRed)
	case ev.Status >= 300:
		statusColor = color.New(color.FgYellow)
	}
	return fmt.Sprintf("%-6s %-30s %s  %s",
		color.CyanString("%s", ev.Method),
		ev.Path,
		statusColor.Sprintf("%d", ev.Status),
		ev.Duration.Round(time.Millisecond),
	)
}
