//go:build !linux

package tracingClient

// Only the standard log package is shipped off Linux; local runs keep stdout untouched.
func teeStdout(*lineShipper) {}
