//go:build windows

package main

import (
	"os"
	"syscall"

	"golang.org/x/sys/windows"
)

var (
	kernel32     = syscall.NewLazyDLL("kernel32.dll")
	allocConsole = kernel32.NewProc("AllocConsole")
)

// ensureConsole allocates a Windows console when the process is launched
// without one (e.g. double-clicked in Explorer). It then reopens stdin,
// stdout, and stderr against the new console so that bubbletea can detect
// a proper terminal and render the TUI correctly.
func ensureConsole() {
	// If stdout is already a terminal we're running inside an existing
	// console (cmd.exe / PowerShell / Windows Terminal) — nothing to do.
	if fileIsTerminal(os.Stdout) {
		return
	}

	// Allocate a new console window.
	allocConsole.Call() //nolint:errcheck

	// Re-open the standard handles so the Go runtime and bubbletea see them.
	reopenConsoleFile("CONIN$", windows.STD_INPUT_HANDLE)
	reopenConsoleFile("CONOUT$", windows.STD_OUTPUT_HANDLE)
	reopenConsoleFile("CONOUT$", windows.STD_ERROR_HANDLE)
}

func fileIsTerminal(f *os.File) bool {
	var mode uint32
	handle := windows.Handle(f.Fd())
	return windows.GetConsoleMode(handle, &mode) == nil
}

func reopenConsoleFile(name string, stdHandle uint32) {
	f, err := os.OpenFile(name, os.O_RDWR, 0)
	if err != nil {
		return
	}
	_ = windows.SetStdHandle(stdHandle, windows.Handle(f.Fd()))
	switch stdHandle {
	case windows.STD_INPUT_HANDLE:
		os.Stdin = f
	case windows.STD_OUTPUT_HANDLE:
		os.Stdout = f
	case windows.STD_ERROR_HANDLE:
		os.Stderr = f
	}
}
