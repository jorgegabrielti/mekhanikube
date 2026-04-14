package k8s

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"testing"

	"k8s.io/client-go/kubernetes/fake"
)

func TestCheckConnection_Success(t *testing.T) {
	client := fake.NewSimpleClientset()
	if err := CheckConnection(context.Background(), client); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestClassifyConnectionError_Auth(t *testing.T) {
	cases := []string{
		"Unauthorized",
		"the server has asked for the client to provide credentials",
		"getting credentials: exec plugin returned non-zero exit code",
		"token has expired",
		"unable to authenticate: security token is expired",
		"AccessDenied: no credentials found",
		"could not get token: InvalidIdentityToken",
	}
	for _, msg := range cases {
		t.Run(truncate(msg, 40), func(t *testing.T) {
			err := classifyConnectionError(errors.New(msg))
			if err == nil {
				t.Fatal("expected error")
			}
			s := err.Error()
			if !strings.Contains(s, "authentication failed") {
				t.Errorf("expected auth error, got: %s", s)
			}
			if !strings.Contains(s, "aws sso login") {
				t.Error("should contain AWS SSO hint")
			}
		})
	}
}

func TestClassifyConnectionError_Network(t *testing.T) {
	cases := []struct {
		name string
		err  error
	}{
		{"connection refused", errors.New("dial tcp 10.0.0.1:443: connection refused")},
		{"no such host", errors.New("dial tcp: lookup cluster.example.com: no such host")},
		{"i/o timeout", errors.New("dial tcp 10.0.0.1:443: i/o timeout")},
		{"net.Error timeout", &net.OpError{Op: "dial", Err: &stubTimeoutErr{}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := classifyConnectionError(tc.err)
			if err == nil {
				t.Fatal("expected error")
			}
			s := err.Error()
			if !strings.Contains(s, "cluster unreachable") {
				t.Errorf("expected network error, got: %s", s)
			}
		})
	}
}

func TestClassifyConnectionError_TLS(t *testing.T) {
	cases := []string{
		"x509: certificate signed by unknown authority",
		"tls: handshake failure",
		"certificate has expired or is not yet valid",
	}
	for _, msg := range cases {
		t.Run(truncate(msg, 40), func(t *testing.T) {
			err := classifyConnectionError(errors.New(msg))
			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.Contains(err.Error(), "TLS/certificate error") {
				t.Errorf("expected TLS error, got: %s", err.Error())
			}
		})
	}
}

func TestClassifyConnectionError_Fallback(t *testing.T) {
	err := classifyConnectionError(errors.New("something unexpected"))
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "failed to connect to cluster") {
		t.Errorf("expected fallback message, got: %s", err.Error())
	}
}

func TestIsNetError(t *testing.T) {
	var target net.Error

	netErr := &net.OpError{Op: "dial", Err: fmt.Errorf("refused")}
	if !isNetError(netErr, &target) {
		t.Error("should detect direct net.Error")
	}

	wrapped := fmt.Errorf("wrap: %w", netErr)
	if !isNetError(wrapped, &target) {
		t.Error("should detect wrapped net.Error")
	}

	if isNetError(errors.New("plain"), &target) {
		t.Error("should not detect plain error as net.Error")
	}
}

// --- helpers ---

type stubTimeoutErr struct{}

func (e *stubTimeoutErr) Error() string   { return "i/o timeout" }
func (e *stubTimeoutErr) Timeout() bool   { return true }
func (e *stubTimeoutErr) Temporary() bool { return true }

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
