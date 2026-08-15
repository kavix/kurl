//go:build darwin
// +build darwin

package client

import (
	"syscall"

	"golang.org/x/sys/unix"
)

func dialerControl(network, address string, c syscall.RawConn) error {
	err := c.Control(func(fd uintptr) {
		// TCP_FASTOPEN is available on macOS
		_ = unix.SetsockoptInt(int(fd), unix.IPPROTO_TCP, unix.TCP_FASTOPEN, 1)
	})
	if err != nil {
		return nil
	}
	// Ignore errors so we fallback to standard TCP if TFO is not supported
	return nil
}
