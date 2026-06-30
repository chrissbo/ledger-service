// Package httpapi exposes the ledger over HTTP. The handler is decoupled
// from the server bootstrap so tests can use httptest without touching
// process-wide state.
package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/chrissbo/ledger-service/internal/ledger"
)

// Version is set at build time via -ldflags.
var Version = "1.2.0"

// New returns an http.Handler that exposes the given ledger.
func New(l *ledger.Ledger) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", healthz)
	mux.HandleFunc("GET /version", version)
	mux.Handle("GET /metrics", MetricsHandler())
	mux.HandleFunc("GET /balance/{account}", instrumentHandler("/balance", balance(l)))
	mux.HandleFunc("POST /deposits", instrumentHandler("/deposits", deposit(l)))
	mux.HandleFunc("POST /transfers", instrumentHandler("/transfers", transfer(l)))
	return mux
}

func healthz(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

func version(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"version": Version})
}

func balance(l *ledger.Ledger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		account := strings.TrimSpace(r.PathValue("account"))
		if account == "" {
			writeErr(w, http.StatusBadRequest, "account required")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"account": account,
			"balance": l.Balance(account),
		})
	}
}

type depositReq struct {
	Account string `json:"account"`
	Amount  int64  `json:"amount"`
}

func deposit(l *ledger.Ledger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req depositReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeErr(w, http.StatusBadRequest, "invalid json")
			return
		}
		if req.Account == "" {
			writeErr(w, http.StatusBadRequest, "account required")
			return
		}
		if err := l.Deposit(req.Account, req.Amount); err != nil {
			writeErr(w, statusForLedgerError(err), err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, map[string]any{
			"account": req.Account,
			"balance": l.Balance(req.Account),
		})
	}
}

type transferReq struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Amount int64  `json:"amount"`
}

func transfer(l *ledger.Ledger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req transferReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeErr(w, http.StatusBadRequest, "invalid json")
			return
		}
		if req.From == "" || req.To == "" {
			writeErr(w, http.StatusBadRequest, "from and to required")
			return
		}
		if err := l.Transfer(req.From, req.To, req.Amount); err != nil {
			writeErr(w, statusForLedgerError(err), err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"from":         req.From,
			"to":           req.To,
			"from_balance": l.Balance(req.From),
			"to_balance":   l.Balance(req.To),
		})
	}
}

func statusForLedgerError(err error) int {
	switch {
	case errors.Is(err, ledger.ErrInvalidAmount):
		return http.StatusBadRequest
	case errors.Is(err, ledger.ErrInsufficientFunds):
		return http.StatusUnprocessableEntity
	default:
		return http.StatusInternalServerError
	}
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeErr(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
