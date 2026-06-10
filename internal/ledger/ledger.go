// Package ledger is an in-memory toy ledger used to exercise the simulated
// CI/CD pipeline. It is intentionally simple — the value of this code is in
// being non-trivial enough to drive lint/test/build checks, not in being a
// real ledger.
package ledger

import (
	"errors"
	"sync"
)

// ErrInsufficientFunds is returned when a transfer would overdraw an account.
var ErrInsufficientFunds = errors.New("insufficient funds")

// ErrInvalidAmount is returned when a non-positive amount is requested.
var ErrInvalidAmount = errors.New("amount must be positive")

// Ledger holds account balances. Safe for concurrent use.
type Ledger struct {
	mu       sync.Mutex
	balances map[string]int64
}

// New returns an empty Ledger.
func New() *Ledger {
	return &Ledger{balances: make(map[string]int64)}
}

// Deposit adds amount to account. Amount must be positive.
func (l *Ledger) Deposit(account string, amount int64) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	l.balances[account] += amount
	return nil
}

// Transfer moves amount from src to dst atomically. Both accounts are
// implicitly created if they don't exist.
func (l *Ledger) Transfer(src, dst string, amount int64) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.balances[src] < amount {
		return ErrInsufficientFunds
	}
	l.balances[src] -= amount
	l.balances[dst] += amount
	return nil
}

// Balance returns the current balance for account (zero for unknown accounts).
func (l *Ledger) Balance(account string) int64 {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.balances[account]
}
