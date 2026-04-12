//go:build windows

package main

import (
	"os"
	"syscall"

	"golang.org/x/sys/windows"
)

var (
	kernel32     = syscall.NewLazyDLL("kernel32.dll") //nolint:gochecknoglobals
	allocConsole = kernel32.NewProc("AllocConsole")   //nolint:gochecknoglobals
)

// ensureConsole allocates a Windows console when the process is launched
// without one (e.g. double-clicked in Explorer). After AllocConsole sets
// up the OS standard handles, we wrap them with os.NewFile so that the Go
// runtime and bubbletea see proper terminal file objects.
func ensureConsole() {
	// If stdout is already a terminal we're running inside an existing
	// console (cmd.exe / PowerShell / Windows Terminal) — nothing to do.
	if fileIsTerminal(os.Stdout) {
		return
	}

	// Allocate a new console window. AllocConsole sets STD_INPUT/OUTPUT/ERROR
	// handles on the process; subsequent GetStdHandle calls return them.
	allocConsole.Call() //nolint:errcheck

	// Wrap the newly set OS handles in Go file objects so that bubbletea's
	// terminal detection (GetConsoleMode on os.Stdout.Fd()) succeeds.
	if h, err := windows.GetStdHandle(windows.STD_INPUT_HANDLE); err == nil {
		os.Stdin = os.NewFile(uintptr(h), "/dev/stdin")
	}
	if h, err := windows.GetStdHandle(windows.STD_OUTPUT_HANDLE); err == nil {
		os.Stdout = os.NewFile(uintptr(h), "/dev/stdout")
	}
	if h, err := windows.GetStdHandle(windows.STD_ERROR_HANDLE); err == nil {
		os.Stderr = os.NewFile(uintptr(h), "/dev/stderr")
	}
}

func fileIsTerminal(f *os.File) bool {
	var mode uint32
	handle := windows.Handle(f.Fd())
	return windows.GetConsoleMode(handle, &mode) == nil
}
