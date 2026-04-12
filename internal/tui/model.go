package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jorgegabrielti/nautikube/internal/diagnosis"
	"github.com/jorgegabrielti/nautikube/internal/k8s"
	"github.com/jorgegabrielti/nautikube/internal/scanner"
	"github.com/jorgegabrielti/nautikube/internal/version"
)

// viewState represents which screen the TUI is on.
type viewState int

const (
	stateSetup viewState = iota
	stateScanning
	stateResults
	stateDetail
	stateOutput
	stateError
)

// scanDoneMsg is sent when the background scan completes.
type scanDoneMsg struct {
	problems []diagnosis.Problem
	err      error
}

// model is the Bubble Tea application model.
type model struct {
	opts     Options
	state    viewState
	spinner  spinner.Model
	problems []diagnosis.Problem
	cursor   int
	err      error
	width    int
	height   int
	// context selection
	contexts    []k8s.ContextInfo
	setupCursor int
	// command output
	cmdTitle   string
	cmdOutput  []string
	cmdScroll  int
	cmdRunning bool
}

func newModel(opts Options, contexts []k8s.ContextInfo) model {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = spinnerStyle

	state := stateScanning
	setupCursor := 0
	// Show context selection only when there are multiple contexts to choose from.
	if len(contexts) > 1 {
		state = stateSetup
		for i, c := range contexts {
			if c.IsCurrent {
				setupCursor = i
				break
			}
		}
	}

	return model{
		opts:        opts,
		state:       state,
		spinner:     sp,
		contexts:    contexts,
		setupCursor: setupCursor,
	}
}

// Init starts the spinner and launches the scan goroutine.
func (m model) Init() tea.Cmd {
	if m.state == stateSetup {
		return nil
	}
	return tea.Batch(m.spinner.Tick, doScan(m.opts))
}

// doScan runs the cluster scan asynchronously and returns a Cmd.
func doScan(opts Options) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		var k8sOpts []k8s.Option
		if opts.Kubeconfig != "" {
			k8sOpts = append(k8sOpts, k8s.WithKubeconfig(opts.Kubeconfig))
		}
		if opts.Context != "" {
			k8sOpts = append(k8sOpts, k8s.WithContext(opts.Context))
		}

		client, err := k8s.NewClient(k8sOpts...)
		if err != nil {
			return scanDoneMsg{err: fmt.Errorf("failed to connect to cluster: %w", err)}
		}

		kb, err := diagnosis.NewKnowledgeBase()
		if err != nil {
			return scanDoneMsg{err: fmt.Errorf("failed to load knowledge base: %w", err)}
		}

		registry := scanner.DefaultRegistry()
		problems, err := registry.ScanAll(ctx, client, opts.Namespace, kb, nil)
		if err != nil {
			return scanDoneMsg{err: fmt.Errorf("scan failed: %w", err)}
		}

		return scanDoneMsg{problems: problems}
	}
}

// Update handles messages and key presses.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case spinner.TickMsg:
		if m.state == stateScanning || m.cmdRunning {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

	case cmdDoneMsg:
		m.cmdRunning = false
		if msg.output == "" {
			m.cmdOutput = []string{"(no output)"}
		} else {
			m.cmdOutput = strings.Split(strings.ReplaceAll(msg.output, "\r\n", "\n"), "\n")
		}
		m.cmdScroll = 0
		return m, nil

	case scanDoneMsg:
		if msg.err != nil {
			m.state = stateError
			m.err = msg.err
		} else {
			m.state = stateResults
			m.problems = msg.problems
			m.cursor = 0
		}
		return m, nil

	case tea.KeyMsg:
		switch m.state {
		case stateSetup:
			return m.updateSetup(msg)
		case stateResults:
			return m.updateResults(msg)
		case stateDetail:
			return m.updateDetail(msg)
		case stateOutput:
			return m.updateOutput(msg)
		case stateError:
			if msg.String() == "q" || msg.String() == "ctrl+c" {
				return m, tea.Quit
			}
			if msg.String() == "r" {
				m.state = stateScanning
				m.err = nil
				return m, tea.Batch(m.spinner.Tick, doScan(m.opts))
			}
			if msg.String() == "c" {
				m.err = nil
				return m.switchContext()
			}
		case stateScanning:
			if msg.String() == "ctrl+c" {
				return m, tea.Quit
			}
		}
	}
	return m, nil
}

func (m model) updateResults(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.problems)-1 {
			m.cursor++
		}
	case "enter", " ":
		if len(m.problems) > 0 {
			m.state = stateDetail
		}
	case "r":
		m.state = stateScanning
		m.problems = nil
		m.cursor = 0
		return m, tea.Batch(m.spinner.Tick, doScan(m.opts))
	case "c":
		return m.switchContext()
	}
	return m, nil
}

func (m model) updateDetail(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "esc", "backspace":
		m.state = stateResults
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.problems)-1 {
			m.cursor++
		}
	case "c":
		return m.switchContext()
	default:
		// Keys 1-9: run the Nth remediation command
		if len(msg.String()) == 1 && msg.String() >= "1" && msg.String() <= "9" {
			idx := int(msg.String()[0] - '1')
			p := m.problems[m.cursor]
			if idx < len(p.Remediation) {
				m.cmdTitle = p.Remediation[idx]
				m.cmdOutput = nil
				m.cmdScroll = 0
				m.cmdRunning = true
				m.state = stateOutput
				return m, tea.Batch(m.spinner.Tick, runCmd(p.Remediation[idx]))
			}
		}
	}
	return m, nil
}

// View renders the current state.
func (m model) View() string {
	switch m.state {
	case stateSetup:
		return m.viewSetup()
	case stateScanning:
		return m.viewScanning()
	case stateResults:
		return m.viewResults()
	case stateDetail:
		return m.viewDetail()
	case stateOutput:
		return m.viewOutput()
	case stateError:
		return m.viewError()
	}
	return ""
}

// switchContext reloads kubeconfig contexts and navigates to the setup screen.
func (m model) switchContext() (tea.Model, tea.Cmd) {
	contexts, _ := k8s.ListContexts()
	if len(contexts) == 0 {
		return m, nil
	}
	m.contexts = contexts
	m.setupCursor = 0
	for i, c := range contexts {
		if c.Name == m.opts.Context || c.IsCurrent {
			m.setupCursor = i
			break
		}
	}
	m.state = stateSetup
	m.problems = nil
	m.cursor = 0
	return m, nil
}

func (m model) updateSetup(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "up", "k":
		if m.setupCursor > 0 {
			m.setupCursor--
		}
	case "down", "j":
		if m.setupCursor < len(m.contexts)-1 {
			m.setupCursor++
		}
	case "enter", " ":
		selected := m.contexts[m.setupCursor]
		m.opts.Context = selected.Name
		m.state = stateScanning
		return m, tea.Batch(m.spinner.Tick, doScan(m.opts))
	}
	return m, nil
}

func (m model) viewSetup() string {
	var sb strings.Builder
	sb.WriteString(m.banner())
	sb.WriteString(titleStyle.Render("  Select a Kubernetes context:"))
	sb.WriteString("\n\n")

	for i, ctx := range m.contexts {
		marker := "  "
		if ctx.IsCurrent {
			marker = infoStyle.Render("★ ")
		}
		cluster := ctx.Cluster
		if cluster == "" {
			cluster = "—"
		}
		row := fmt.Sprintf("%s%-40s %s", marker, truncate(ctx.Name, 40), subtitleStyle.Render("cluster: "+truncate(cluster, 30)))
		if i == m.setupCursor {
			sb.WriteString(selectedStyle.Render(row))
		} else {
			sb.WriteString(normalRowStyle.Render(row))
		}
		sb.WriteString("\n")
	}

	sb.WriteString(helpStyle.Render("  ↑/↓ navigate  enter select  q quit"))
	return sb.String()
}

func (m model) viewScanning() string {
	ns := m.opts.Namespace
	if ns == "" {
		ns = "all namespaces"
	}
	return fmt.Sprintf("%s\n  %s Scanning cluster (%s)...\n\n  %s",
		m.banner(),
		m.spinner.View(),
		ns,
		helpStyle.Render("ctrl+c to quit"),
	)
}

func (m model) viewResults() string {
	var sb strings.Builder

	// Header
	ns := m.opts.Namespace
	if ns == "" {
		ns = "all namespaces"
	}
	sb.WriteString(m.banner())
	sb.WriteString(subtitleStyle.Render(fmt.Sprintf("Namespace: %s  |  Problems found: %d", ns, len(m.problems))))
	sb.WriteString("\n\n")

	if len(m.problems) == 0 {
		sb.WriteString(infoStyle.Render("  No problems detected. Cluster looks healthy!"))
		sb.WriteString("\n")
	} else {
		// Column header
		header := fmt.Sprintf("  %-10s %-12s %-22s %-5s  %s",
			"SEVERITY", "RESOURCE", "NAME", "SCORE", "ISSUE")
		sb.WriteString(subtitleStyle.Render(header))
		sb.WriteString("\n")
		sb.WriteString(subtitleStyle.Render("  " + strings.Repeat("─", 80)))
		sb.WriteString("\n")

		// Rows
		visibleStart, visibleEnd := m.visibleRange()
		for i := visibleStart; i < visibleEnd; i++ {
			p := m.problems[i]
			sevStr := m.colorSeverity(p.Severity)
			name := truncate(p.Name, 22)
			issue := truncate(p.Issue, 40)
			row := fmt.Sprintf("  %-10s %-12s %-22s %-5d  %s",
				sevStr, p.Resource, name, p.Score, issue)

			if i == m.cursor {
				sb.WriteString(selectedStyle.Render(row))
			} else {
				sb.WriteString(normalRowStyle.Render(row))
			}
			sb.WriteString("\n")
		}

		// Scroll hint
		if len(m.problems) > m.visibleRows() {
			sb.WriteString(subtitleStyle.Render(fmt.Sprintf("\n  Showing %d-%d of %d",
				visibleStart+1, visibleEnd, len(m.problems))))
			sb.WriteString("\n")
		}
	}

	sb.WriteString(helpStyle.Render("  ↑/↓ navigate  enter detail  r rescan  c context  q quit"))
	return sb.String()
}

func (m model) viewDetail() string {
	if len(m.problems) == 0 || m.cursor >= len(m.problems) {
		return ""
	}
	p := m.problems[m.cursor]

	var sb strings.Builder
	sb.WriteString(m.banner())
	sb.WriteString(subtitleStyle.Render("Problem Detail"))
	sb.WriteString("\n")

	field := func(key, val string) {
		sb.WriteString("  ")
		sb.WriteString(detailKeyStyle.Render(key + ":"))
		sb.WriteString(" ")
		sb.WriteString(detailValStyle.Render(val))
		sb.WriteString("\n")
	}

	field("Severity", string(p.Severity))
	field("Resource", p.Resource)
	field("Namespace", p.Namespace)
	field("Name", p.Name)
	field("Score", fmt.Sprintf("%d", p.Score))
	field("Issue", p.Issue)

	if p.OffendingProperty != "" {
		field("Offending Property", p.OffendingProperty)
	}
	if p.Explanation != "" {
		sb.WriteString("\n  ")
		sb.WriteString(detailKeyStyle.Render("Explanation:"))
		sb.WriteString("\n")
		for _, line := range wrapText(p.Explanation, m.width-6) {
			sb.WriteString("    ")
			sb.WriteString(detailValStyle.Render(line))
			sb.WriteString("\n")
		}
	}
	if len(p.Remediation) > 0 {
		sb.WriteString("\n  ")
		sb.WriteString(detailKeyStyle.Render("Remediation:"))
		sb.WriteString("\n")
		for i, cmd := range p.Remediation {
			label := fmt.Sprintf("[%d]", i+1)
			sb.WriteString("    ")
			sb.WriteString(detailKeyStyle.Render(label))
			sb.WriteString(" ")
			sb.WriteString(detailValStyle.Render(cmd))
			sb.WriteString("\n")
		}
	}
	if p.MutativeFix != "" {
		sb.WriteString("\n  ")
		sb.WriteString(detailKeyStyle.Render("Mutative Fix:"))
		sb.WriteString("\n    ")
		sb.WriteString(detailValStyle.Render(p.MutativeFix))
		sb.WriteString("\n")
	}
	if len(p.Details) > 0 {
		sb.WriteString("\n  ")
		sb.WriteString(detailKeyStyle.Render("Details:"))
		sb.WriteString("\n")
		for _, d := range p.Details {
			sb.WriteString("    ")
			sb.WriteString(detailValStyle.Render("• " + d))
			sb.WriteString("\n")
		}
	}

	nav := fmt.Sprintf("  Problem %d of %d", m.cursor+1, len(m.problems))
	sb.WriteString("\n")
	sb.WriteString(subtitleStyle.Render(nav))
	sb.WriteString("\n")
	sb.WriteString(helpStyle.Render("  ↑/↓ next/prev  1-9 run cmd  esc back  c context  r rescan  q quit"))
	return sb.String()
}

func (m model) updateOutput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "esc", "backspace":
		m.state = stateDetail
	case "up", "k":
		if m.cmdScroll > 0 {
			m.cmdScroll--
		}
	case "down", "j":
		if m.cmdScroll < len(m.cmdOutput)-1 {
			m.cmdScroll++
		}
	}
	return m, nil
}

func (m model) viewOutput() string {
	var sb strings.Builder
	sb.WriteString(m.banner())

	if m.cmdRunning {
		sb.WriteString("\n  ")
		sb.WriteString(m.spinner.View())
		sb.WriteString(" Running...\n\n  ")
		sb.WriteString(detailValStyle.Render(m.cmdTitle))
		sb.WriteString("\n")
		return sb.String()
	}

	sb.WriteString(subtitleStyle.Render("Command Output"))
	sb.WriteString("\n  ")
	sb.WriteString(detailKeyStyle.Render("$"))
	sb.WriteString(" ")
	sb.WriteString(detailValStyle.Render(m.cmdTitle))
	sb.WriteString("\n\n")

	// visible lines
	chrome := m.bannerLines() + 6
	visible := m.height - chrome
	if visible < 1 {
		visible = 1
	}

	end := m.cmdScroll + visible
	if end > len(m.cmdOutput) {
		end = len(m.cmdOutput)
	}
	for _, line := range m.cmdOutput[m.cmdScroll:end] {
		sb.WriteString("  ")
		sb.WriteString(line)
		sb.WriteString("\n")
	}

	sb.WriteString("\n")
	sb.WriteString(helpStyle.Render("  ↑/↓ scroll  esc back  q quit"))
	return sb.String()
}

func (m model) viewError() string {
	var sb strings.Builder
	sb.WriteString(m.banner())
	sb.WriteString(errorStyle.Render("  Error running scan:"))
	sb.WriteString("\n\n  ")
	sb.WriteString(detailValStyle.Render(m.err.Error()))
	sb.WriteString("\n\n")
	sb.WriteString(helpStyle.Render("  r retry  c context  q quit"))
	return sb.String()
}

// visibleRows returns how many rows fit given the terminal height and UI chrome.
func (m model) visibleRows() int {
	// chrome = banner lines + subtitle(1) + blank(1) + col-header(1) + divider(1) + scroll-hint(1) + help(1)
	chrome := m.bannerLines() + 6
	rows := m.height - chrome
	if rows < 5 {
		return 5
	}
	return rows
}

// bannerLines returns the number of terminal lines the banner occupies.
func (m model) bannerLines() int {
	return strings.Count(m.banner(), "\n")
}

func (m model) visibleRange() (int, int) {
	total := len(m.problems)
	rows := m.visibleRows()
	start := m.cursor - rows/2
	if start < 0 {
		start = 0
	}
	end := start + rows
	if end > total {
		end = total
		start = end - rows
		if start < 0 {
			start = 0
		}
	}
	return start, end
}

// colorSeverity styles a severity string.
func (m model) colorSeverity(sev diagnosis.Severity) string {
	s := string(sev)
	switch sev {
	case diagnosis.Critical:
		return criticalStyle.Render(s)
	case diagnosis.High:
		return highStyle.Render(s)
	case diagnosis.Medium:
		return mediumStyle.Render(s)
	case diagnosis.Low:
		return lowStyle.Render(s)
	default:
		return infoStyle.Render(s)
	}
}

// truncate shortens a string to max length, appending "…" if needed.
func truncate(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max-1]) + "…"
}

// wrapText breaks text into lines of at most width runes.
func wrapText(text string, width int) []string {
	if width <= 0 {
		width = 80
	}
	words := strings.Fields(text)
	var lines []string
	var current strings.Builder
	for _, w := range words {
		if current.Len()+len(w)+1 > width {
			lines = append(lines, current.String())
			current.Reset()
		}
		if current.Len() > 0 {
			current.WriteByte(' ')
		}
		current.WriteString(w)
	}
	if current.Len() > 0 {
		lines = append(lines, current.String())
	}
	if len(lines) == 0 {
		return []string{text}
	}
	return lines
}

// banner renders the ASCII art logo + version line.
// Falls back to a compact one-liner when the terminal is narrow.
func (m model) banner() string {
	const minWidth = 72
	if m.width > 0 && m.width < minWidth {
		return bannerStyle.Render("NautiKube") + "  " + versionBadgeStyle.Render("v"+version.Version) + "\n\n"
	}
	art := `
  ███╗   ██╗ █████╗ ██╗   ██╗████████╗██╗██╗  ██╗██╗   ██╗██████╗ ███████╗
  ████╗  ██║██╔══██╗██║   ██║╚══██╔══╝██║██║ ██╔╝██║   ██║██╔══██╗██╔════╝
  ██╔██╗ ██║███████║██║   ██║   ██║   ██║█████╔╝ ██║   ██║██████╔╝█████╗
  ██║╚██╗██║██╔══██║██║   ██║   ██║   ██║██╔═██╗ ██║   ██║██╔══██╗██╔══╝
  ██║ ╚████║██║  ██║╚██████╔╝   ██║   ██║██║  ██╗╚██████╔╝██████╔╝███████╗
  ╚═╝  ╚═══╝╚═╝  ╚═╝ ╚═════╝    ╚═╝   ╚═╝╚═╝  ╚═╝ ╚═════╝ ╚═════╝ ╚══════╝`
	return bannerStyle.Render(art) + "\n" + versionBadgeStyle.Render("  v"+version.Version+"  —  Kubernetes Cluster Diagnostic Tool") + "\n\n"
}

// ensure lipgloss import is used (borderStyle is available for future use)
var _ = lipgloss.NewStyle()
var _ = borderStyle
