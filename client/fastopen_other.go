//go:build !linux && !darwin
// +build !linux,!darwin

package client

import (
	"syscall"
)

func dialerControl(network, address string, c syscall.RawConn) error {
	// TCP Fast Open not supported or enabled on other platforms
	return nil
}
