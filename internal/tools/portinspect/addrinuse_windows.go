//go:build windows

package portinspect

import (
	"errors"
	"syscall"
)

const windowsAddressInUse syscall.Errno = 10048
const windowsPermissionDenied syscall.Errno = 10013

func isAddressInUse(err error) bool {
	return errors.Is(err, windowsAddressInUse)
}

func isPermissionDenied(err error) bool {
	return errors.Is(err, windowsPermissionDenied)
}
