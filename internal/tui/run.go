package tui

import (
tea "github.com/charmbracelet/bubbletea"
"github.com/jorgegabrielti/nautikube/internal/k8s"
)

// Options holds configuration for the TUI session.
type Options struct {
Namespace  string
Kubeconfig string
Context    string
}

// Run starts the interactive TUI mode.
func Run(opts Options) error {
// Load kubeconfig contexts synchronously (fast local file read).
// If --context was already specified via CLI, skip the selection screen.
var contexts []k8s.ContextInfo
if opts.Context == "" {
contexts, _ = k8s.ListContexts()
}

m := newModel(opts, contexts)
p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
_, err := p.Run()
return err
}
