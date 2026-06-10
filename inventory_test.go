package test

import (
	"sync"
	"sync/atomic"
	"testing"
)

func TestReserve_ConcurrentOversell(t *testing.T) {
	products := map[string]*Product{
		"p1": {ID: "p1", Name: "Product", Stock: 100},
	}
	svc := NewSafeInventoryService(products)

	var wg sync.WaitGroup
	var successCount atomic.Int32
	var failCount atomic.Int32

	for i := 0; i < 200; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := svc.Reserve("p1", 1)
			if err == nil {
				successCount.Add(1)
			} else {
				failCount.Add(1)
			}
		}()
	}

	wg.Wait()

	if successCount.Load() != 100 {
		t.Errorf("expected 100 successes, got %d", successCount.Load())
	}
	if failCount.Load() != 100 {
		t.Errorf("expected 100 failures, got %d", failCount.Load())
	}
	if svc.GetStock("p1") != 0 {
		t.Errorf("expected stock 0, got %d", svc.GetStock("p1"))
	}
}

func TestReserveMultiple_Atomicity(t *testing.T) {
	products := map[string]*Product{
		"A": {ID: "A", Stock: 10},
		"B": {ID: "B", Stock: 5},
	}
	svc := NewSafeInventoryService(products)

	err := svc.ReserveMultiple([]ReserveItem{
		{ProductID: "A", Quantity: 8},
		{ProductID: "B", Quantity: 8},
	})
	if err != ErrInsufficientStock {
		t.Fatalf("expected ErrInsufficientStock, got %v", err)
	}
	if svc.GetStock("A") != 10 {
		t.Errorf("expected A stock 10, got %d", svc.GetStock("A"))
	}
	if svc.GetStock("B") != 5 {
		t.Errorf("expected B stock 5, got %d", svc.GetStock("B"))
	}
}
