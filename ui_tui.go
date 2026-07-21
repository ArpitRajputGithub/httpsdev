package main

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const feedSize = 20

type tickMsg time.Time
type eventMsg Event
type doneMsg struct{}

type model struct {
	listenAddr string
	upstream   string
	feed       []Event
	total      int
	errors     int
	totalDur   time.Duration
	startedAt  time.Time
	tab        int // 0 = requests, 1 = cert info
}

func initialModel(listenAddr, upstream string) model {
	return model{
		listenAddr: listenAddr,
		upstream:   upstream,
		startedAt:  time.Now(),
	}
}

func (m model) Init() tea.Cmd { return tea.Batch(tick(), tea.EnterAltScreen) }

func tick() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg { return tickMsg(t) })
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "c":
			m.feed = nil
		case "1":
			m.tab = 0
		case "2":
			m.tab = 1
		}
	case tickMsg:
		return m, tick()
	case eventMsg:
		ev := Event(msg)
		m.total++
		m.totalDur += ev.Duration
		if ev.Status >= 400 {
			m.errors++
		}
		m.feed = append(m.feed, ev)
		if len(m.feed) > feedSize {
			m.feed = m.feed[len(m.feed)-feedSize:]
		}
	case doneMsg:
		return m, tea.Quit
	}
	return m, nil
}

// Palette adapted from torlink (github.com/baairon/torlink).
var (
	accent = lipgloss.Color("#a78bfa") // violet — active/focus/title
	muted  = lipgloss.Color("#6b6577") // purple-grey — dividers, labels, inactive
	ok     = lipgloss.Color("#86d6a2") // soft green — 2xx / good counts
	warn   = lipgloss.Color("#f0c674") // 3xx
	bad    = lipgloss.Color("#e57373") // 4xx/5xx / error counts

	titleStyle  = lipgloss.NewStyle().Bold(true).Foreground(accent)
	labelStyle  = lipgloss.NewStyle().Foreground(muted)
	dimStyle    = lipgloss.NewStyle().Foreground(muted).Faint(true)
	sectionHdr  = lipgloss.NewStyle().Foreground(accent).Bold(true)
	borderStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(muted).Padding(0, 1)
	statOK      = lipgloss.NewStyle().Foreground(ok)
	statBad     = lipgloss.NewStyle().Foreground(bad)
	statWarn    = lipgloss.NewStyle().Foreground(warn)
	tabActive   = lipgloss.NewStyle().Foreground(accent).Bold(true)
	tabInactive = lipgloss.NewStyle().Foreground(muted)
	chev        = lipgloss.NewStyle().Foreground(accent).Render("❯")
)

func (m model) View() string {
	var avg time.Duration
	if m.total > 0 {
		avg = m.totalDur / time.Duration(m.total)
	}
	uptime := time.Since(m.startedAt).Round(time.Second)

	header := chev + "  " + titleStyle.Render("httpsdev v"+Version) + "   " +
		labelStyle.Render(m.listenAddr+"  →  "+m.upstream)

	statsHdr := sectionHdr.Render("▲ live")
	statsBox := borderStyle.Render(fmt.Sprintf(
		"total  %d    errors  %s    avg  %s    uptime  %s",
		m.total,
		colorInt(m.errors, m.errors == 0),
		avg.Round(time.Millisecond),
		uptime,
	))

	var mainHdr, mainBox string
	switch m.tab {
	case 0:
		mainHdr = sectionHdr.Render("▲ requests")
		mainBox = borderStyle.Render(m.renderFeed())
	case 1:
		mainHdr = sectionHdr.Render("▲ cert")
		mainBox = borderStyle.Render(strings.Join([]string{
			"cert     " + labelStyle.Render("mkcert-signed"),
			"subject  " + labelStyle.Render("localhost"),
			"valid    " + labelStyle.Render("30 days (rolling)"),
			"",
			dimStyle.Render("more introspection in v0.2"),
		}, "\n"))
	}

	tabs := []string{"1 requests", "2 cert"}
	if m.tab == 0 {
		tabs[0] = tabActive.Render(tabs[0])
		tabs[1] = tabInactive.Render(tabs[1])
	} else {
		tabs[0] = tabInactive.Render(tabs[0])
		tabs[1] = tabActive.Render(tabs[1])
	}
	footer := strings.Join(tabs, "   ") + labelStyle.Render("      q quit · c clear")

	return strings.Join([]string{"", header, "", statsHdr, statsBox, "", mainHdr, mainBox, "", footer}, "\n")
}

func (m model) renderFeed() string {
	if len(m.feed) == 0 {
		return dimStyle.Render("waiting for requests…")
	}
	var b strings.Builder
	lastIdx := len(m.feed) - 1
	for i, ev := range m.feed {
		s := statOK
		switch {
		case ev.Status >= 500:
			s = statBad
		case ev.Status >= 400:
			s = statBad
		case ev.Status >= 300:
			s = statWarn
		}
		marker := "  "
		if i == lastIdx {
			marker = chev + " "
		}
		fmt.Fprintf(&b, "%s%-6s %-30s %s  %s\n",
			marker,
			ev.Method,
			truncate(ev.Path, 30),
			s.Render(fmt.Sprintf("%d", ev.Status)),
			labelStyle.Render(ev.Duration.Round(time.Millisecond).String()),
		)
	}
	return strings.TrimRight(b.String(), "\n")
}

func colorInt(n int, ok bool) string {
	if ok {
		return statOK.Render(fmt.Sprintf("%d", n))
	}
	return statBad.Render(fmt.Sprintf("%d", n))
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}

// runTUI drives the bubbletea program, feeding it events from the channel.
func runTUI(listenAddr, upstream string, events <-chan Event) {
	p := tea.NewProgram(initialModel(listenAddr, upstream))
	go func() {
		for ev := range events {
			p.Send(eventMsg(ev))
		}
		p.Send(doneMsg{})
	}()
	_, _ = p.Run()
}
