//go:build unit

package ledger

import (
	"errors"
	"sync"
	"testing"
)

func TestDeposit(t *testing.T) {
	t.Parallel()
	l := New()
	if err := l.Deposit("alice", 100); err != nil {
		t.Fatalf("Deposit returned unexpected error: %v", err)
	}
	if got, want := l.Balance("alice"), int64(100); got != want {
		t.Fatalf("Balance(alice) = %d; want %d", got, want)
	}
}

func TestDepositRejectsNonPositive(t *testing.T) {
	t.Parallel()
	l := New()
	for _, amt := range []int64{0, -1, -100} {
		if err := l.Deposit("alice", amt); !errors.Is(err, ErrInvalidAmount) {
			t.Fatalf("Deposit(%d) error = %v; want ErrInvalidAmount", amt, err)
		}
	}
}

func TestTransferHappyPath(t *testing.T) {
	t.Parallel()
	l := New()
	mustNoErr(t, l.Deposit("alice", 500))
	if err := l.Transfer("alice", "bob", 200); err != nil {
		t.Fatalf("Transfer returned unexpected error: %v", err)
	}
	if got := l.Balance("alice"); got != 300 {
		t.Fatalf("alice balance = %d; want 300", got)
	}
	if got := l.Balance("bob"); got != 200 {
		t.Fatalf("bob balance = %d; want 200", got)
	}
}

func TestTransferInsufficientFunds(t *testing.T) {
	t.Parallel()
	l := New()
	mustNoErr(t, l.Deposit("alice", 50))
	err := l.Transfer("alice", "bob", 100)
	if !errors.Is(err, ErrInsufficientFunds) {
		t.Fatalf("Transfer error = %v; want ErrInsufficientFunds", err)
	}
	if got := l.Balance("alice"); got != 50 {
		t.Fatalf("alice balance after failed transfer = %d; want 50", got)
	}
	if got := l.Balance("bob"); got != 0 {
		t.Fatalf("bob balance after failed transfer = %d; want 0", got)
	}
}

// TestConcurrentTransfers is the reason go test runs with -race in CI.
func TestConcurrentTransfers(t *testing.T) {
	t.Parallel()
	l := New()
	mustNoErr(t, l.Deposit("alice", 10000))

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = l.Transfer("alice", "bob", 1)
		}()
	}
	wg.Wait()

	if got := l.Balance("alice") + l.Balance("bob"); got != 10000 {
		t.Fatalf("conservation broken: total = %d; want 10000", got)
	}
}

func mustNoErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
