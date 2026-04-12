package tui

import (
	"bytes"
	"os/exec"
	"runtime"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// cmdDoneMsg carries the result of an executed command.
type cmdDoneMsg struct {
	output string
	err    error
}

// runCmd executes cmdStr via the OS shell and returns a tea.Cmd.
// Inline comments (# ...) at the end of the command are stripped first.
func runCmd(cmdStr string) tea.Cmd {
	return func() tea.Msg {
		clean := stripComment(cmdStr)
		var cmd *exec.Cmd
		if runtime.GOOS == "windows" {
			cmd = exec.Command("powershell.exe", "-NoProfile", "-NonInteractive", "-Command", clean)
		} else {
			cmd = exec.Command("sh", "-c", clean)
		}

		var out bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &out

		_ = cmd.Run() // intentionally ignore exit code — show output regardless
		return cmdDoneMsg{output: out.String()}
	}
}

// stripComment removes a trailing `# …` inline comment from a shell command,
// but only outside of quoted segments (simple heuristic: cut at first ` #`).
func stripComment(cmd string) string {
	if idx := strings.Index(cmd, " #"); idx != -1 {
		return strings.TrimSpace(cmd[:idx])
	}
	return strings.TrimSpace(cmd)
}
