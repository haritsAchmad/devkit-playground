//go:build windows

package portinspect

import (
	"fmt"
	"syscall"
	"testing"
)

func TestWindowsSocketErrorClassification(t *testing.T) {
	if !isAddressInUse(fmt.Errorf("wrapped: %w", syscall.Errno(10048))) {
		t.Error("Winsock WSAEADDRINUSE was not classified as address in use")
	}
	if !isPermissionDenied(fmt.Errorf("wrapped: %w", syscall.Errno(10013))) {
		t.Error("Winsock WSAEACCES was not classified as permission denied")
	}
}
