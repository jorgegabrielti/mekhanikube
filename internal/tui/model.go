package tui

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jorgegabrielti/nautikube/internal/config"
	"github.com/jorgegabrielti/nautikube/internal/diagnosis"
	"github.com/jorgegabrielti/nautikube/internal/k8s"
	"github.com/jorgegabrielti/nautikube/internal/scanner"
	"github.com/jorgegabrielti/nautikube/internal/version"
	"k8s.io/client-go/kubernetes"
)

// viewState represents which screen the TUI is on.
type viewState int

const (
	stateInitConfig viewState = iota
	stateSetup
	stateConnecting
	stateScanning
	stateResults
	stateDetail
	stateOutput
	stateReportMenu
	stateReportFormat
	stateReportSaved
	stateError
)

// connectDoneMsg is sent when the cluster connectivity check completes.
type connectDoneMsg struct {
	client kubernetes.Interface
	err    error
}

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
	errStep  string // which checklist step failed
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
	// report
	reportOrigin       viewState
	reportPath         string
	reportMenuCursor   int // 0=full, 1=current issue (scope selection)
	reportScope        int // confirmed scope: 0=full, 1=current issue
	reportFormatCursor int // 0=TXT, 1=CSV, 2=PDF
	// init config
	initStep        int // 0=lang, 1=severity, 2=output, 3=retention
	initCursor      int
	initLang        string
	initSev         string
	initOutput      string
	initRetention   int       // history retention in days
	initReturnState viewState // screen to return to after reconfiguration
}

func newModel(opts Options, contexts []k8s.ContextInfo) model {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = spinnerStyle

	state := stateConnecting
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

	// If no config file exists, show initial configuration screen first.
	cfgPath, _ := config.Path()
	needInit := false
	if cfgPath != "" {
		if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
			needInit = true
		}
	}

	m := model{
		opts:          opts,
		state:         state,
		spinner:       sp,
		contexts:      contexts,
		setupCursor:   setupCursor,
		initLang:      "en",
		initSev:       "",
		initOutput:    "table",
		initRetention: 90,
	}
	if needInit {
		m.state = stateInitConfig
	}
	return m
}

// Init starts the spinner and launches the connect check.
func (m model) Init() tea.Cmd {
	if m.state == stateSetup || m.state == stateInitConfig {
		return nil
	}
	return tea.Batch(m.spinner.Tick, doConnect(m.opts))
}

// doConnect checks cluster connectivity and returns the client.
func doConnect(opts Options) tea.Cmd {
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
			return connectDoneMsg{err: fmt.Errorf("failed to connect to cluster: %w", err)}
		}

		if err := k8s.CheckConnection(ctx, client); err != nil {
			return connectDoneMsg{err: err}
		}

		return connectDoneMsg{client: client}
	}
}

// doScan runs the cluster scan using an already-connected client.
func doScan(opts Options, client kubernetes.Interface) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		kb, err := diagnosis.NewKnowledgeBase(opts.Lang)
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
		if m.state == stateConnecting || m.state == stateScanning || m.cmdRunning {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

	case connectDoneMsg:
		if msg.err != nil {
			m.state = stateError
			m.err = msg.err
			m.errStep = "connect"
		} else {
			m.state = stateScanning
			return m, doScan(m.opts, msg.client)
		}
		return m, nil

	case cmdDoneMsg:
		m.cmdRunning = false
		if msg.output == "" {
			m.cmdOutput = []string{t(m.opts.Lang, "no_output")}
		} else {
			m.cmdOutput = strings.Split(strings.ReplaceAll(msg.output, "\r\n", "\n"), "\n")
		}
		m.cmdScroll = 0
		return m, nil

	case scanDoneMsg:
		if msg.err != nil {
			m.state = stateError
			m.err = msg.err
			m.errStep = "scan"
		} else {
			m.state = stateResults
			m.problems = msg.problems
			m.cursor = 0
		}
		return m, nil

	case tea.KeyMsg:
		switch m.state {
		case stateInitConfig:
			return m.updateInitConfig(msg)
		case stateSetup:
			return m.updateSetup(msg)
		case stateResults:
			return m.updateResults(msg)
		case stateDetail:
			return m.updateDetail(msg)
		case stateOutput:
			return m.updateOutput(msg)
		case stateReportMenu:
			return m.updateReportMenu(msg)
		case stateReportFormat:
			return m.updateReportFormat(msg)
		case stateReportSaved:
			m.state = m.reportOrigin
			return m, nil
		case stateError:
			if msg.String() == "q" || msg.String() == "ctrl+c" {
				return m, tea.Quit
			}
			if msg.String() == "r" {
				m.state = stateConnecting
				m.err = nil
				m.errStep = ""
				return m, tea.Batch(m.spinner.Tick, doConnect(m.opts))
			}
			if msg.String() == "c" {
				m.err = nil
				return m.switchContext()
			}
		case stateConnecting, stateScanning:
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
		m.state = stateConnecting
		m.problems = nil
		m.cursor = 0
		m.errStep = ""
		return m, tea.Batch(m.spinner.Tick, doConnect(m.opts))
	case "e":
		if len(m.problems) > 0 {
			m.reportOrigin = stateResults
			m.reportMenuCursor = 0
			m.state = stateReportMenu
		}
	case "c":
		return m.switchContext()
	case "s":
		m.initStep = 0
		m.initCursor = 0
		cfg := config.Load()
		m.initLang = cfg.Language
		m.initSev = cfg.Severity
		m.initOutput = cfg.Output
		m.initRetention = cfg.HistoryRetention
		m.initReturnState = stateResults
		m.state = stateInitConfig
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
	case "e":
		m.reportOrigin = stateDetail
		m.reportMenuCursor = 0
		m.state = stateReportMenu
	case "c":
		return m.switchContext()
	case "s":
		m.initStep = 0
		m.initCursor = 0
		cfg := config.Load()
		m.initLang = cfg.Language
		m.initSev = cfg.Severity
		m.initOutput = cfg.Output
		m.initRetention = cfg.HistoryRetention
		m.initReturnState = stateDetail
		m.state = stateInitConfig
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
	case stateInitConfig:
		return m.viewInitConfig()
	case stateSetup:
		return m.viewSetup()
	case stateConnecting, stateScanning:
		return m.viewProgress()
	case stateResults:
		return m.viewResults()
	case stateDetail:
		return m.viewDetail()
	case stateOutput:
		return m.viewOutput()
	case stateReportMenu:
		return m.viewReportMenu()
	case stateReportFormat:
		return m.viewReportFormat()
	case stateReportSaved:
		return m.viewReportSaved()
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

func (m model) updateInitConfig(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	options := m.initOptions()
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "up", "k":
		if m.initCursor > 0 {
			m.initCursor--
		}
	case "down", "j":
		if m.initCursor < len(options)-1 {
			m.initCursor++
		}
	case "enter", " ":
		sel := options[m.initCursor]
		switch m.initStep {
		case 0:
			m.initLang = sel
			m.initStep = 1
			m.initCursor = 0
		case 1:
			m.initSev = sel
			m.initStep = 2
			m.initCursor = 0
		case 2:
			m.initOutput = sel
			m.initStep = 3
			m.initCursor = 0
			// Pre-select current retention option.
			retOpts := m.initOptions()
			for i, o := range retOpts {
				if o == fmt.Sprintf("%d", m.initRetention) {
					m.initCursor = i
					break
				}
			}
		case 3:
			// Parse retention value.
			switch sel {
			case "0":
				m.initRetention = 0
			case "30":
				m.initRetention = 30
			case "90":
				m.initRetention = 90
			case "180":
				m.initRetention = 180
			case "365":
				m.initRetention = 365
			case "-1":
				m.initRetention = -1
			}
			// Save config and proceed.
			cfg := config.Config{
				Language:         m.initLang,
				Severity:         m.initSev,
				Output:           m.initOutput,
				HistoryRetention: m.initRetention,
			}
			_ = config.Save(cfg)
			// Apply to opts so downstream uses the chosen language.
			m.opts.Lang = m.initLang
			// Return to previous screen if reconfiguring, otherwise proceed.
			if m.initReturnState != stateInitConfig {
				ret := m.initReturnState
				m.initReturnState = stateInitConfig
				m.state = ret
				if ret == stateResults || ret == stateDetail {
					// Re-scan with new language.
					m.state = stateConnecting
					m.problems = nil
					m.cursor = 0
					return m, tea.Batch(m.spinner.Tick, doConnect(m.opts))
				}
				return m, nil
			}
			// First-time setup: proceed to context selection or connect.
			if len(m.contexts) > 1 {
				m.state = stateSetup
				return m, nil
			}
			m.state = stateConnecting
			return m, tea.Batch(m.spinner.Tick, doConnect(m.opts))
		}
	}
	return m, nil
}

func (m model) initOptions() []string {
	switch m.initStep {
	case 0:
		return []string{"en", "pt"}
	case 1:
		return []string{"", "critical", "high", "medium", "low"}
	case 2:
		return []string{"table", "json", "yaml"}
	case 3:
		return []string{"0", "30", "90", "180", "365", "-1"}
	}
	return nil
}

// retentionLabel returns a human-readable label for a retention option value.
func (m model) retentionLabel(lang, val string) string {
	switch val {
	case "0":
		return t(lang, "ret_disabled")
	case "30":
		return t(lang, "ret_30")
	case "90":
		return t(lang, "ret_90")
	case "180":
		return t(lang, "ret_180")
	case "365":
		return t(lang, "ret_365")
	case "-1":
		return t(lang, "ret_unlimited")
	}
	return val
}

func (m model) viewInitConfig() string {
	l := m.opts.Lang
	var sb strings.Builder
	sb.WriteString(m.banner())
	sb.WriteString(titleStyle.Render("  " + t(l, "init_title")))
	sb.WriteString("\n\n")

	stepLabels := []string{t(l, "step_lang"), t(l, "step_severity"), t(l, "step_output"), t(l, "step_retention")}
	sb.WriteString(subtitleStyle.Render(fmt.Sprintf("  Step %d/4: %s", m.initStep+1, stepLabels[m.initStep])))
	sb.WriteString("\n\n")

	options := m.initOptions()
	for i, opt := range options {
		label := opt
		if label == "" {
			label = t(l, "sev_all")
		}
		// Retention step: show human-readable labels.
		if m.initStep == 3 {
			label = m.retentionLabel(l, opt)
		}
		if i == m.initCursor {
			sb.WriteString(selectedStyle.Render(fmt.Sprintf("  ▸ %s", label)))
		} else {
			sb.WriteString(normalRowStyle.Render(fmt.Sprintf("    %s", label)))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("\n")
	sb.WriteString(helpStyle.Render(t(l, "help_init")))
	return sb.String()
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
		m.state = stateConnecting
		return m, tea.Batch(m.spinner.Tick, doConnect(m.opts))
	case "s":
		m.initStep = 0
		m.initCursor = 0
		cfg := config.Load()
		m.initLang = cfg.Language
		m.initSev = cfg.Severity
		m.initOutput = cfg.Output
		m.initRetention = cfg.HistoryRetention
		m.initReturnState = stateSetup
		m.state = stateInitConfig
	}
	return m, nil
}

func (m model) viewSetup() string {
	l := m.opts.Lang
	var sb strings.Builder
	sb.WriteString(m.banner())
	sb.WriteString(titleStyle.Render("  " + t(l, "setup_title")))
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

	sb.WriteString(helpStyle.Render(t(l, "help_setup")))
	return sb.String()
}

func (m model) viewProgress() string {
	l := m.opts.Lang
	var sb strings.Builder
	sb.WriteString(m.banner())

	ctxName := m.opts.Context
	if ctxName == "" {
		ctxName = t(l, "current_context")
	}
	ns := m.opts.Namespace
	if ns == "" {
		ns = t(l, "all_namespaces")
	}

	switch m.state {
	case stateConnecting:
		sb.WriteString(fmt.Sprintf("  %s %s (%s)...\n", m.spinner.View(), t(l, "connecting"), ctxName))
		sb.WriteString(subtitleStyle.Render(fmt.Sprintf("    %s (%s)", t(l, "scanning_res"), ns)))
		sb.WriteString("\n")
	case stateScanning:
		sb.WriteString(infoStyle.Render("  ✓"))
		sb.WriteString(fmt.Sprintf(" %s (%s)\n", t(l, "connected"), ctxName))
		sb.WriteString(fmt.Sprintf("  %s %s (%s)...\n", m.spinner.View(), t(l, "scanning_res"), ns))
	}

	sb.WriteString("\n")
	sb.WriteString(helpStyle.Render("  ctrl+c"))
	return sb.String()
}

func (m model) viewResults() string {
	l := m.opts.Lang
	var sb strings.Builder

	// Header
	ns := m.opts.Namespace
	if ns == "" {
		ns = t(l, "all_namespaces")
	}
	sb.WriteString(m.banner())
	sb.WriteString(subtitleStyle.Render(fmt.Sprintf("%s: %s  |  %s: %d", t(l, "ns_label"), ns, t(l, "problems_found"), len(m.problems))))
	sb.WriteString("\n\n")

	if len(m.problems) == 0 {
		sb.WriteString(infoStyle.Render("  " + t(l, "no_problems")))
		sb.WriteString("\n")
	} else {
		// Column header
		header := fmt.Sprintf("  %-10s %-12s %-22s %-5s  %s",
			t(l, "col_severity"), t(l, "col_resource"), t(l, "col_name"), t(l, "col_score"), t(l, "col_issue"))
		sb.WriteString(subtitleStyle.Render(header))
		sb.WriteString("\n")
		sb.WriteString(subtitleStyle.Render("  " + strings.Repeat("─", 80)))
		sb.WriteString("\n")

		// Rows
		visibleStart, visibleEnd := m.visibleRange()
		for i := visibleStart; i < visibleEnd; i++ {
			p := m.problems[i]
			// Pad plain text FIRST, then apply color — avoids ANSI byte-count misalignment.
			sevPadded := fmt.Sprintf("%-10s", string(p.Severity))
			sevCol := m.colorSeverityStr(p.Severity, sevPadded)
			row := fmt.Sprintf("  %s %-12s %-22s %-5d  %s",
				sevCol,
				truncate(p.Resource, 12),
				truncate(p.Name, 22),
				p.Score,
				truncate(p.Issue, 40))

			if i == m.cursor {
				sb.WriteString(selectedStyle.Render(row))
			} else {
				sb.WriteString(normalRowStyle.Render(row))
			}
			sb.WriteString("\n")
		}

		// Scroll hint
		if len(m.problems) > m.visibleRows() {
			sb.WriteString(subtitleStyle.Render(fmt.Sprintf("\n  %s %d-%d %s %d",
				t(l, "showing"), visibleStart+1, visibleEnd, t(l, "of"), len(m.problems))))
			sb.WriteString("\n")
		}
	}

	sb.WriteString(helpStyle.Render(t(l, "help_results")))
	return sb.String()
}

func (m model) viewDetail() string {
	l := m.opts.Lang
	if len(m.problems) == 0 || m.cursor >= len(m.problems) {
		return ""
	}
	p := m.problems[m.cursor]

	var sb strings.Builder
	sb.WriteString(m.banner())
	sb.WriteString(subtitleStyle.Render(t(l, "detail_title")))
	sb.WriteString("\n")

	field := func(key, val string) {
		sb.WriteString("  ")
		sb.WriteString(detailKeyStyle.Render(key + ":"))
		sb.WriteString(" ")
		sb.WriteString(detailValStyle.Render(val))
		sb.WriteString("\n")
	}

	field(t(l, "field_severity"), string(p.Severity))
	field(t(l, "field_resource"), p.Resource)
	field(t(l, "field_namespace"), p.Namespace)
	field(t(l, "field_name"), p.Name)
	field(t(l, "field_score"), fmt.Sprintf("%d", p.Score))
	field(t(l, "field_issue"), p.Issue)

	if p.OffendingProperty != "" {
		field(t(l, "field_offending"), p.OffendingProperty)
	}
	if p.Explanation != "" {
		sb.WriteString("\n  ")
		sb.WriteString(detailKeyStyle.Render(t(l, "field_explanation") + ":"))
		sb.WriteString("\n")
		for _, line := range wrapText(p.Explanation, m.width-6) {
			sb.WriteString("    ")
			sb.WriteString(detailValStyle.Render(line))
			sb.WriteString("\n")
		}
	}
	if len(p.Remediation) > 0 {
		sb.WriteString("\n  ")
		sb.WriteString(detailKeyStyle.Render(t(l, "field_remediation") + ":"))
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
	if len(p.Details) > 0 {
		sb.WriteString("\n  ")
		sb.WriteString(detailKeyStyle.Render(t(l, "field_details") + ":"))
		sb.WriteString("\n")
		for _, d := range p.Details {
			sb.WriteString("    ")
			sb.WriteString(detailValStyle.Render("• " + d))
			sb.WriteString("\n")
		}
	}

	nav := fmt.Sprintf("  "+t(l, "problem_x_of_y"), m.cursor+1, len(m.problems))
	sb.WriteString("\n")
	sb.WriteString(subtitleStyle.Render(nav))
	sb.WriteString("\n")
	sb.WriteString(helpStyle.Render(t(l, "help_detail")))
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
	l := m.opts.Lang
	var sb strings.Builder
	sb.WriteString(m.banner())

	if m.cmdRunning {
		sb.WriteString("\n  ")
		sb.WriteString(m.spinner.View())
		sb.WriteString(" " + t(l, "running") + "\n\n  ")
		sb.WriteString(detailValStyle.Render(m.cmdTitle))
		sb.WriteString("\n")
		return sb.String()
	}

	sb.WriteString(subtitleStyle.Render(t(l, "cmd_output")))
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
	sb.WriteString(helpStyle.Render(t(l, "help_output")))
	return sb.String()
}

func (m model) updateReportMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "esc", "backspace":
		m.state = m.reportOrigin
	case "up", "k":
		if m.reportMenuCursor > 0 {
			m.reportMenuCursor--
		}
	case "down", "j":
		if m.reportMenuCursor < 1 {
			m.reportMenuCursor++
		}
	case "enter", " ":
		m.reportScope = m.reportMenuCursor
		m.reportFormatCursor = 0
		m.state = stateReportFormat
	}
	return m, nil
}

func (m model) updateReportFormat(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "esc", "backspace":
		m.state = stateReportMenu
	case "up", "k":
		if m.reportFormatCursor > 0 {
			m.reportFormatCursor--
		}
	case "down", "j":
		if m.reportFormatCursor < 2 {
			m.reportFormatCursor++
		}
	case "enter", " ":
		var path string
		var err error
		switch m.reportScope {
		case 0: // full report
			switch m.reportFormatCursor {
			case 0:
				path, err = saveFullReport(m.problems, m.opts.Context, m.opts.Namespace)
			case 1:
				path, err = saveCSVReport(m.problems, m.opts.Context, m.opts.Namespace)
			case 2:
				path, err = savePDFReport(m.problems, m.opts.Context, m.opts.Namespace)
			}
		case 1: // current issue
			if len(m.problems) > 0 && m.cursor < len(m.problems) {
				switch m.reportFormatCursor {
				case 0:
					path, err = saveIssueReport(m.problems[m.cursor], m.opts.Context)
				case 1:
					path, err = saveCSVIssueReport(m.problems[m.cursor], m.opts.Context)
				case 2:
					path, err = savePDFIssueReport(m.problems[m.cursor], m.opts.Context)
				}
			}
		}
		if err != nil {
			m.reportPath = "Error: " + err.Error()
		} else {
			m.reportPath = path
		}
		m.state = stateReportSaved
	}
	return m, nil
}

func (m model) viewReportFormat() string {
	l := m.opts.Lang
	var sb strings.Builder
	sb.WriteString(m.banner())
	scopeLabel := t(l, "full_report")
	if m.reportScope == 1 {
		scopeLabel = t(l, "current_issue")
	}
	sb.WriteString(subtitleStyle.Render(fmt.Sprintf("%s — %s", t(l, "export_format"), scopeLabel)))
	sb.WriteString("\n\n")

	formats := []string{
		t(l, "txt_desc"),
		t(l, "csv_desc"),
		t(l, "pdf_desc"),
	}
	for i, f := range formats {
		label := "  [ ] "
		if i == m.reportFormatCursor {
			label = "  [●] "
		}
		row := label + f
		if i == m.reportFormatCursor {
			sb.WriteString(selectedStyle.Render(row))
		} else {
			sb.WriteString(normalRowStyle.Render(row))
		}
		sb.WriteString("\n")
	}

	sb.WriteString(helpStyle.Render(t(l, "help_report")))
	return sb.String()
}

func (m model) viewReportMenu() string {
	l := m.opts.Lang
	var sb strings.Builder
	sb.WriteString(m.banner())
	sb.WriteString(subtitleStyle.Render(t(l, "export_report")))
	sb.WriteString("\n\n")

	options := []string{
		fmt.Sprintf("%s  (%d %s)", t(l, "full_report"), len(m.problems), t(l, "issues")),
		t(l, "current_issue"),
	}

	for i, opt := range options {
		label := "  [ ] "
		if i == m.reportMenuCursor {
			label = "  [●] "
		}
		row := label + opt
		if i == m.reportMenuCursor {
			sb.WriteString(selectedStyle.Render(row))
		} else {
			sb.WriteString(normalRowStyle.Render(row))
		}
		sb.WriteString("\n")
	}

	sb.WriteString(helpStyle.Render(t(l, "help_report")))
	return sb.String()
}

func (m model) viewReportSaved() string {
	l := m.opts.Lang
	var sb strings.Builder
	sb.WriteString(m.banner())
	if strings.HasPrefix(m.reportPath, "Error:") {
		sb.WriteString(errorStyle.Render("  " + t(l, "export_failed")))
		sb.WriteString("\n\n  ")
		sb.WriteString(detailValStyle.Render(m.reportPath))
	} else {
		sb.WriteString(infoStyle.Render("  " + t(l, "report_saved")))
		sb.WriteString("\n\n  ")
		sb.WriteString(detailValStyle.Render(m.reportPath))
	}
	sb.WriteString("\n\n")
	sb.WriteString(helpStyle.Render("  " + t(l, "any_key")))
	return sb.String()
}

func (m model) viewError() string {
	l := m.opts.Lang
	var sb strings.Builder
	sb.WriteString(m.banner())

	ctxName := m.opts.Context
	if ctxName == "" {
		ctxName = t(l, "current_context")
	}
	ns := m.opts.Namespace
	if ns == "" {
		ns = t(l, "all_namespaces")
	}

	switch m.errStep {
	case "connect":
		sb.WriteString(errorStyle.Render("  ✗"))
		sb.WriteString(fmt.Sprintf(" %s (%s)\n", t(l, "conn_failed"), ctxName))
		sb.WriteString(subtitleStyle.Render(fmt.Sprintf("    %s (%s)", t(l, "scanning_res"), ns)))
		sb.WriteString("\n")
	case "scan":
		sb.WriteString(infoStyle.Render("  ✓"))
		sb.WriteString(fmt.Sprintf(" %s (%s)\n", t(l, "connected"), ctxName))
		sb.WriteString(errorStyle.Render("  ✗"))
		sb.WriteString(fmt.Sprintf(" %s (%s)\n", t(l, "scan_failed"), ns))
	default:
		sb.WriteString(errorStyle.Render("  ✗ " + t(l, "error")))
		sb.WriteString("\n")
	}

	sb.WriteString("\n  ")
	sb.WriteString(detailValStyle.Render(m.err.Error()))
	sb.WriteString("\n\n")
	sb.WriteString(helpStyle.Render(t(l, "help_error")))
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

// colorSeverity styles a severity string (uses the raw severity text).
func (m model) colorSeverity(sev diagnosis.Severity) string {
	return m.colorSeverityStr(sev, string(sev))
}

// colorSeverityStr applies the severity color to an arbitrary string s.
// Use this when s is pre-padded so ANSI codes don't break column alignment.
func (m model) colorSeverityStr(sev diagnosis.Severity, s string) string {
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
