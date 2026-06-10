//go:build integration

// This integration test runs against a live HTTP server bound to a real
// loopback port. In a real service this would also stand up testcontainers
// for postgres, kafka, etc. — for the simulation, the in-process server is
// enough to demonstrate that the integration build tag is wired correctly.

package httpapi_test

import (
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/chrissbo/ledger-service/internal/httpapi"
	"github.com/chrissbo/ledger-service/internal/ledger"
)

func TestEndToEndTransferFlow(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	srv := &http.Server{
		Handler:           httpapi.New(ledger.New()),
		ReadHeaderTimeout: 2 * time.Second,
	}
	go func() { _ = srv.Serve(listener) }()
	t.Cleanup(func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	})

	base := "http://" + listener.Addr().String()
	client := &http.Client{Timeout: 5 * time.Second}

	// Deposit 1000 to alice.
	mustPost(t, client, base+"/deposits", `{"account":"alice","amount":1000}`, http.StatusCreated)

	// Transfer 400 alice -> bob.
	mustPost(t, client, base+"/transfers", `{"from":"alice","to":"bob","amount":400}`, http.StatusOK)

	// Verify balances.
	if got := getBalance(t, client, base, "alice"); got != 600 {
		t.Fatalf("alice balance = %d; want 600", got)
	}
	if got := getBalance(t, client, base, "bob"); got != 400 {
		t.Fatalf("bob balance = %d; want 400", got)
	}
}

func mustPost(t *testing.T, client *http.Client, url, body string, wantStatus int) {
	t.Helper()
	resp, err := client.Post(url, "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("POST %s: %v", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != wantStatus {
		out, _ := io.ReadAll(resp.Body)
		t.Fatalf("POST %s status = %d; want %d; body=%s", url, resp.StatusCode, wantStatus, out)
	}
}

func getBalance(t *testing.T, client *http.Client, base, account string) int64 {
	t.Helper()
	resp, err := client.Get(base + "/balance/" + account)
	if err != nil {
		t.Fatalf("GET balance: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET balance status = %d", resp.StatusCode)
	}
	var body struct {
		Balance int64 `json:"balance"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	return body.Balance
}
