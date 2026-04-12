//go:build !windows

package main

// ensureConsole is a no-op on non-Windows platforms.
func ensureConsole() {}
