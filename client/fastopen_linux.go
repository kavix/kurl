//go:build linux
// +build linux

package client

import (
	"syscall"

	"golang.org/x/sys/unix"
)

func dialerControl(network, address string, c syscall.RawConn) error {
	err := c.Control(func(fd uintptr) {
		// TCP_FASTOPEN_CONNECT is available on modern Linux
		_ = unix.SetsockoptInt(int(fd), unix.SOL_TCP, unix.TCP_FASTOPEN_CONNECT, 1)
	})
	if err != nil {
		return nil
	}
	// Ignore errors so we fallback to standard TCP if TFO is not supported
	return nil
}
