//go:build unit

package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/chrissbo/ledger-service/internal/ledger"
)

func TestHealthz(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(New(ledger.New()))
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/healthz")
	if err != nil {
		t.Fatalf("GET /healthz: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d; want 200", resp.StatusCode)
	}
}

func TestDepositThenBalance(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(New(ledger.New()))
	defer srv.Close()

	body := strings.NewReader(`{"account":"alice","amount":250}`)
	resp, err := http.Post(srv.URL+"/deposits", "application/json", body)
	if err != nil {
		t.Fatalf("POST /deposits: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d; want 201", resp.StatusCode)
	}

	resp, err = http.Get(srv.URL + "/balance/alice")
	if err != nil {
		t.Fatalf("GET /balance/alice: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d; want 200", resp.StatusCode)
	}
}

func TestTransferInsufficient422(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(New(ledger.New()))
	defer srv.Close()

	body := strings.NewReader(`{"from":"alice","to":"bob","amount":100}`)
	resp, err := http.Post(srv.URL+"/transfers", "application/json", body)
	if err != nil {
		t.Fatalf("POST /transfers: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d; want 422", resp.StatusCode)
	}
}

func TestDepositRejectsBadJSON(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(New(ledger.New()))
	defer srv.Close()

	resp, err := http.Post(srv.URL+"/deposits", "application/json", strings.NewReader(`{not json`))
	if err != nil {
		t.Fatalf("POST /deposits: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d; want 400", resp.StatusCode)
	}
}
