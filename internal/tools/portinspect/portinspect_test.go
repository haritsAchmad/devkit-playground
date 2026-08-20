package portinspect

import (
	"errors"
	"net"
	"testing"
)

func TestInspectReportsPortInUse(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen fixture: %v", err)
	}
	defer listener.Close()
	port := listener.Addr().(*net.TCPAddr).Port

	result, err := Inspect("127.0.0.1", port)
	if err != nil {
		t.Fatalf("Inspect() error = %v", err)
	}
	if result.Host != "127.0.0.1" || result.Port != port || result.Protocol != "tcp" || result.State != StateInUse {
		t.Errorf("result = %+v, want occupied TCP port", result)
	}
}

func TestInspectReportsAvailablePort(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("reserve fixture: %v", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	if err := listener.Close(); err != nil {
		t.Fatalf("release fixture: %v", err)
	}

	result, err := Inspect("127.0.0.1", port)
	if err != nil {
		t.Fatalf("Inspect() error = %v", err)
	}
	if result.State != StateAvailable {
		t.Errorf("State = %q, want available", result.State)
	}
}

func TestInspectRejectsInvalidInput(t *testing.T) {
	if _, err := Inspect("localhost", 8080); !errors.Is(err, ErrInvalidHost) {
		t.Errorf("Inspect() host error = %v, want ErrInvalidHost", err)
	}
	for _, port := range []int{0, 65536} {
		if _, err := Inspect("127.0.0.1", port); !errors.Is(err, ErrInvalidPort) {
			t.Errorf("Inspect() port %d error = %v, want ErrInvalidPort", port, err)
		}
	}
}
