package main

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const feedSize = 15

// toilet -f mono12 HTTPSDEV
const logo = ` ▄▄    ▄▄  ▄▄▄▄▄▄▄▄  ▄▄▄▄▄▄▄▄  ▄▄▄▄▄▄      ▄▄▄▄    ▄▄▄▄▄     ▄▄▄▄▄▄▄▄  ▄▄    ▄▄
 ██    ██  ▀▀▀██▀▀▀  ▀▀▀██▀▀▀  ██▀▀▀▀█▄  ▄█▀▀▀▀█   ██▀▀▀██   ██▀▀▀▀▀▀  ▀██  ██▀
 ██    ██     ██        ██     ██    ██  ██▄       ██    ██  ██         ██  ██
 ████████     ██        ██     ██████▀    ▀████▄   ██    ██  ███████    ██  ██
 ██    ██     ██        ██     ██             ▀██  ██    ██  ██          ████
 ██    ██     ██        ██     ██        █▄▄▄▄▄█▀  ██▄▄▄██   ██▄▄▄▄▄▄    ████
 ▀▀    ▀▀     ▀▀        ▀▀     ▀▀         ▀▀▀▀▀    ▀▀▀▀▀     ▀▀▀▀▀▀▀▀    ▀▀▀▀`

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
	width      int
}

func initialModel(listenAddr, upstream string) model {
	return model{
		listenAddr: listenAddr,
		upstream:   upstream,
		startedAt:  time.Now(),
		width:      100,
	}
}

func (m model) Init() tea.Cmd { return tea.Batch(tick(), tea.EnterAltScreen) }

func tick() tea.Cmd {
	return tea.Tick(200*time.Millisecond, func(t time.Time) tea.Msg { return tickMsg(t) })
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "c":
			m.feed = nil
			m.total = 0
			m.errors = 0
			m.totalDur = 0
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

// Catppuccin Mocha palette.
var (
	accent = lipgloss.Color("#cba6f7") // mauve
	muted  = lipgloss.Color("#6c7086") // overlay0
	ok     = lipgloss.Color("#a6e3a1") // green
	warn   = lipgloss.Color("#fab387") // peach
	bad    = lipgloss.Color("#f38ba8") // red

	logoStyle   = lipgloss.NewStyle().Foreground(accent).Bold(true)
	dimStyle    = lipgloss.NewStyle().Foreground(muted).Faint(true)
	mutedStyle  = lipgloss.NewStyle().Foreground(muted)
	accentStyle = lipgloss.NewStyle().Foreground(accent).Bold(true)
	statOK      = lipgloss.NewStyle().Foreground(ok)
	statBad     = lipgloss.NewStyle().Foreground(bad)
	statWarn    = lipgloss.NewStyle().Foreground(warn)
	arrowStyle  = lipgloss.NewStyle().Foreground(muted)
	chevStyle   = lipgloss.NewStyle().Foreground(accent).Bold(true)
)

func (m model) View() string {
	w := m.width
	if w < 80 {
		w = 80
	}

	// 1. Wordmark centered + tagline
	logoLines := strings.Split(logo, "\n")
	logoWidth := 0
	for _, l := range logoLines {
		if lipgloss.Width(l) > logoWidth {
			logoWidth = lipgloss.Width(l)
		}
	}
	pad := (w - logoWidth) / 2
	if pad < 0 {
		pad = 0
	}
	padStr := strings.Repeat(" ", pad)
	var logoBlock strings.Builder
	for _, l := range logoLines {
		logoBlock.WriteString(padStr + logoStyle.Render(l) + "\n")
	}
	tagline := "Real browser-trusted HTTPS for any local dev server."
	taglinePad := (w - len(tagline)) / 2
	if taglinePad < 0 {
		taglinePad = 0
	}
	logoBlock.WriteString(strings.Repeat(" ", taglinePad) + dimStyle.Render(tagline) + "\n")

	// 2. Connection strip + cert info
	connLine := "  " + accentStyle.Render("▶") + "  " +
		accentStyle.Render("https://"+m.listenAddr) + "     " +
		arrowStyle.Render("⇢") + "     " +
		mutedStyle.Render(m.upstream)
	certLine := "     " + dimStyle.Render("mkcert · valid 30d")

	// 3. Live section header with inline stats + trailing rule
	var avg time.Duration
	if m.total > 0 {
		avg = m.totalDur / time.Duration(m.total)
	}
	uptime := time.Since(m.startedAt).Round(time.Second)
	statsText := fmt.Sprintf("%d req · %s err · %s avg · %s",
		m.total,
		colorInt(m.errors, m.errors == 0),
		avg.Round(time.Millisecond),
		uptime,
	)
	hdrPrefix := mutedStyle.Render("  ── ") + accentStyle.Render("live") + mutedStyle.Render(" ──────── ") + statsText + " "
	hdrFill := w - lipgloss.Width(hdrPrefix) - 2
	if hdrFill < 0 {
		hdrFill = 0
	}
	sectionHdr := hdrPrefix + mutedStyle.Render(strings.Repeat("─", hdrFill))

	// 4. Request feed
	feed := m.renderFeed()

	// 5. Bottom rule
	bottomRule := "  " + mutedStyle.Render(strings.Repeat("─", w-4))

	// 6. Keybar
	keybar := "  " + dimStyle.Render("[c]") + mutedStyle.Render(" clear    ") +
		dimStyle.Render("[q]") + mutedStyle.Render(" quit")

	return "\n" + logoBlock.String() + "\n" +
		connLine + "\n" +
		certLine + "\n\n" +
		sectionHdr + "\n\n" +
		feed + "\n\n" +
		bottomRule + "\n\n" +
		keybar
}

func (m model) renderFeed() string {
	if len(m.feed) == 0 {
		return "    " + dimStyle.Render("waiting for requests…")
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
		marker := "    " + arrowStyle.Render("→") + " "
		if i == lastIdx {
			marker = "    " + chevStyle.Render("❯") + " "
		}
		fmt.Fprintf(&b, "%s%-6s %-36s %s   %s\n",
			marker,
			ev.Method,
			truncate(ev.Path, 36),
			s.Render(fmt.Sprintf("%d", ev.Status)),
			mutedStyle.Render(ev.Duration.Round(time.Millisecond).String()),
		)
	}
	return strings.TrimRight(b.String(), "\n")
}

func colorInt(n int, isOK bool) string {
	if isOK {
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
