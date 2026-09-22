package idgen

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestGenerator_GenerateOrderNo(t *testing.T) {
	g := New()
	no := g.GenerateOrderNo()

	if len(no) < 2 || no[:2] != "TO" {
		t.Fatalf("expected prefix TO, got %s", no)
	}
	if len(no) != 20 {
		t.Fatalf("expected length 20, got %d: %s", len(no), no)
	}

	expectedDate := time.Now().Format("20060102")
	actualDate := no[2:10]
	if actualDate != expectedDate {
		t.Fatalf("expected date %s, got %s", expectedDate, actualDate)
	}
}

func TestGenerator_GeneratePaymentNo(t *testing.T) {
	g := New()
	no := g.GeneratePaymentNo()

	if len(no) < 2 || no[:2] != "PW" {
		t.Fatalf("expected prefix PW, got %s", no)
	}
	if len(no) != 20 {
		t.Fatalf("expected length 20, got %d: %s", len(no), no)
	}
}

func TestGenerator_GenerateCallbackNo(t *testing.T) {
	g := New()
	no := g.GenerateCallbackNo()

	if len(no) < 2 || no[:2] != "CB" {
		t.Fatalf("expected prefix CB, got %s", no)
	}
	if len(no) != 20 {
		t.Fatalf("expected length 20, got %d: %s", len(no), no)
	}
}

func TestGenerator_SequenceMonotonic(t *testing.T) {
	g := New()

	prev := g.GenerateOrderNo()
	if prev == "" {
		t.Fatal("expected non-empty order no")
	}

	for i := 0; i < 100; i++ {
		next := g.GenerateOrderNo()
		if next == prev {
			t.Fatalf("duplicate order no at iteration %d: %s", i, next)
		}
		prevSeq := prev[10:]
		nextSeq := next[10:]
		if nextSeq <= prevSeq {
			t.Fatalf("sequence not monotonic: prev=%s next=%s", prevSeq, nextSeq)
		}
		prev = next
	}
}

func TestGenerator_ConcurrentUniqueness(t *testing.T) {
	g := New()
	const goroutines = 1000
	results := make(chan string, goroutines)
	var wg sync.WaitGroup

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			results <- g.GenerateOrderNo()
		}()
	}
	wg.Wait()
	close(results)

	seen := make(map[string]bool, goroutines)
	for no := range results {
		if seen[no] {
			t.Fatalf("duplicate order no in concurrent test: %s", no)
		}
		seen[no] = true
	}

	if len(seen) != goroutines {
		t.Fatalf("expected %d unique order nos, got %d", goroutines, len(seen))
	}
}

func TestGenerator_DateChange(t *testing.T) {
	g := New()

	g.date = "20200101"
	g.counter = 9999999999

	next := g.GenerateOrderNo()
	expectedDate := time.Now().Format("20060102")
	actualDate := next[2:10]
	if actualDate != expectedDate {
		t.Fatalf("expected date to reset to %s, got %s", expectedDate, actualDate)
	}

	seq := next[10:]
	if seq != fmt.Sprintf("%010d", 1) {
		t.Fatalf("expected counter to reset to 1, got %s", seq)
	}
}

func TestGenerator_ZeroPadding(t *testing.T) {
	g := New()
	g.date = time.Now().Format("20060102")
	g.counter = 0

	no := g.GenerateOrderNo()
	seq := no[10:]
	if seq != "0000000001" {
		t.Fatalf("expected zero-padded sequence 0000000001, got %s", seq)
	}
}

func TestGenerator_MixedPrefixes(t *testing.T) {
	g := New()

	orderNo := g.GenerateOrderNo()
	paymentNo := g.GeneratePaymentNo()
	callbackNo := g.GenerateCallbackNo()

	if orderNo == paymentNo || paymentNo == callbackNo || orderNo == callbackNo {
		t.Fatal("different prefixes should produce different numbers")
	}

	if orderNo[:2] != "TO" {
		t.Fatalf("expected TO prefix: %s", orderNo)
	}
	if paymentNo[:2] != "PW" {
		t.Fatalf("expected PW prefix: %s", paymentNo)
	}
	if callbackNo[:2] != "CB" {
		t.Fatalf("expected CB prefix: %s", callbackNo)
	}
}