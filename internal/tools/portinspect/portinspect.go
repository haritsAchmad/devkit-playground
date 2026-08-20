// Package portinspect probes local TCP port availability without connecting to a service.
package portinspect

import (
	"errors"
	"fmt"
	"net"
	"strconv"
)

var (
	ErrInvalidHost = errors.New("invalid host")
	ErrInvalidPort = errors.New("invalid port")
	ErrProbeFailed = errors.New("port probe failed")
)

type State string

const (
	StateAvailable   State = "available"
	StateInUse       State = "in_use"
	StateUnavailable State = "unavailable"
)

type Result struct {
	Host     string
	Port     int
	Protocol string
	State    State
	Reason   string
}

// Inspect attempts a short-lived local bind to determine TCP port availability.
func Inspect(host string, port int) (Result, error) {
	ip := net.ParseIP(host)
	if ip == nil {
		return Result{}, fmt.Errorf("%w: must be an IPv4 or IPv6 address", ErrInvalidHost)
	}
	if port < 1 || port > 65535 {
		return Result{}, fmt.Errorf("%w: must be between 1 and 65535", ErrInvalidPort)
	}

	result := Result{Host: ip.String(), Port: port, Protocol: "tcp", State: StateAvailable}
	listener, err := net.Listen("tcp", net.JoinHostPort(result.Host, strconv.Itoa(port)))
	if err != nil {
		if isAddressInUse(err) {
			result.State = StateInUse
			return result, nil
		}
		if isPermissionDenied(err) {
			result.State = StateUnavailable
			result.Reason = "bind_access_denied"
			return result, nil
		}
		return Result{}, fmt.Errorf("%w: %v", ErrProbeFailed, err)
	}
	if err := listener.Close(); err != nil {
		return Result{}, fmt.Errorf("%w: close probe listener: %v", ErrProbeFailed, err)
	}
	return result, nil
}
