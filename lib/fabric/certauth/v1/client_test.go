package certauth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

// TestEnrollWithCSR_ContextCancellation verifies the activity's context reaches
// the enrollment HTTP request: a context canceled before the call aborts the
// request before it is sent. Before the http.NewRequestWithContext fix the
// request ignored ctx entirely, so cancellation/deadlines could not stop it.
func TestEnrollWithCSR_ContextCancellation(t *testing.T) {
	// Contain any fabric-ca client Init() artifacts (keystore/config) in a temp dir.
	t.Chdir(t.TempDir())

	var handlerHit atomic.Bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		handlerHit.Store(true)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client, err := NewBootstrapClient(BootstrapClientConfig{
		URL:      srv.URL,
		Username: "admin",
		Password: "secret",
	})
	if err != nil {
		t.Fatalf("NewBootstrapClient: %v", err)
	}
	defer client.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // canceled before the enrollment call

	csr := []byte("-----BEGIN CERTIFICATE REQUEST-----\nfake\n-----END CERTIFICATE REQUEST-----")
	_, _, err = client.EnrollWithCSR(ctx, csr, "", nil)

	if err == nil {
		t.Fatal("expected an error from the canceled context, got nil")
	}
	if !errors.Is(err, ErrEnroll) {
		t.Errorf("expected error wrapping ErrEnroll, got: %v", err)
	}
	if handlerHit.Load() {
		t.Error("enrollment request reached the server despite the canceled context (ctx not threaded into the request)")
	}
}
