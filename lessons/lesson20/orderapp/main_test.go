package main

import (
	"context"
	"testing"
	"time"
)

func newTestStore() *Store {
	s := NewStore()
	s.products = map[int64]*Product{
		1: {ID: 1, Name: "Keyboard", Stock: 5},
	}
	return s
}

func TestCreateOrder_Idempotent(t *testing.T) {
	ctx := context.Background()
	s := newTestStore()

	input := CreateOrderInput{
		ProductID:      1,
		Quantity:       2,
		IdempotencyKey: "key-1",
		ExpireAfter:    time.Minute,
	}

	first, repeated, err := s.CreateOrder(ctx, input)
	if err != nil {
		t.Fatalf("first create failed: %v", err)
	}
	if repeated {
		t.Fatalf("first create should not be repeated")
	}
	if got := s.products[1].Stock; got != 3 {
		t.Fatalf("stock after first = %d, want 3", got)
	}

	second, repeated, err := s.CreateOrder(ctx, input)
	if err != nil {
		t.Fatalf("second create failed: %v", err)
	}
	if !repeated {
		t.Fatalf("second create should be repeated")
	}
	if second.ID != first.ID {
		t.Fatalf("repeated order id = %d, want %d", second.ID, first.ID)
	}
	if got := s.products[1].Stock; got != 3 {
		t.Fatalf("stock after second = %d, want still 3 (only deduct once)", got)
	}
}

func TestCreateOrder_StockNotEnough(t *testing.T) {
	ctx := context.Background()
	s := newTestStore()

	_, _, err := s.CreateOrder(ctx, CreateOrderInput{
		ProductID:      1,
		Quantity:       999,
		IdempotencyKey: "key-x",
		ExpireAfter:    time.Minute,
	})
	if err == nil {
		t.Fatalf("expected stock error, got nil")
	}
}

func TestCancelExpired(t *testing.T) {
	ctx := context.Background()
	s := newTestStore()

	order, _, err := s.CreateOrder(ctx, CreateOrderInput{
		ProductID:      1,
		Quantity:       2,
		IdempotencyKey: "key-expire",
		ExpireAfter:    50 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	time.Sleep(200 * time.Millisecond)

	cancelled, err := s.CancelExpired(ctx)
	if err != nil {
		t.Fatalf("cancel expired failed: %v", err)
	}
	if cancelled != 1 {
		t.Fatalf("cancelled = %d, want 1", cancelled)
	}
	if got := s.orders[order.ID].Status; got != StatusCancelled {
		t.Fatalf("status = %s, want cancelled", got)
	}
	if got := s.products[1].Stock; got != 5 {
		t.Fatalf("stock after cancel = %d, want 5 (returned)", got)
	}
}

func TestCancelExpired_PaidNotCancelled(t *testing.T) {
	ctx := context.Background()
	s := newTestStore()

	order, _, err := s.CreateOrder(ctx, CreateOrderInput{
		ProductID:      1,
		Quantity:       2,
		IdempotencyKey: "key-paid",
		ExpireAfter:    50 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	if _, err = s.PayOrder(ctx, order.ID); err != nil {
		t.Fatalf("pay failed: %v", err)
	}

	time.Sleep(200 * time.Millisecond)

	cancelled, err := s.CancelExpired(ctx)
	if err != nil {
		t.Fatalf("cancel expired failed: %v", err)
	}
	if cancelled != 0 {
		t.Fatalf("cancelled = %d, want 0 (paid must not be cancelled)", cancelled)
	}
	if got := s.orders[order.ID].Status; got != StatusPaid {
		t.Fatalf("status = %s, want paid", got)
	}
	if got := s.products[1].Stock; got != 3 {
		t.Fatalf("stock = %d, want 3 (never returned)", got)
	}
}

func TestCancelOrder_Pending(t *testing.T) {
	ctx := context.Background()
	s := newTestStore()

	order, _, err := s.CreateOrder(ctx, CreateOrderInput{
		ProductID:      1,
		Quantity:       2,
		IdempotencyKey: "key-cancel",
		ExpireAfter:    time.Minute,
	})
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	got, cancelled, err := s.CancelOrder(ctx, order.ID)
	if err != nil {
		t.Fatalf("cancel failed: %v", err)
	}
	if !cancelled {
		t.Fatalf("cancelled = false, want true")
	}
	if got.Status != StatusCancelled {
		t.Fatalf("status = %s, want cancelled", got.Status)
	}
	if s.products[1].Stock != 5 {
		t.Fatalf("stock = %d, want 5 (returned)", s.products[1].Stock)
	}
}

func TestCancelOrder_Paid(t *testing.T) {
	ctx := context.Background()
	s := newTestStore()

	order, _, err := s.CreateOrder(ctx, CreateOrderInput{
		ProductID:      1,
		Quantity:       2,
		IdempotencyKey: "key-paid-cancel",
		ExpireAfter:    time.Minute,
	})
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if _, err = s.PayOrder(ctx, order.ID); err != nil {
		t.Fatalf("pay failed: %v", err)
	}

	_, _, err = s.CancelOrder(ctx, order.ID)
	if err == nil {
		t.Fatalf("expected error for paid order, got nil")
	}
	if s.products[1].Stock != 3 {
		t.Fatalf("stock = %d, want 3 (never returned)", s.products[1].Stock)
	}
}

func TestCancelOrder_AlreadyCancelled(t *testing.T) {
	ctx := context.Background()
	s := newTestStore()

	order, _, err := s.CreateOrder(ctx, CreateOrderInput{
		ProductID:      1,
		Quantity:       2,
		IdempotencyKey: "key-cancel-twice",
		ExpireAfter:    time.Minute,
	})
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	if _, _, err = s.CancelOrder(ctx, order.ID); err != nil {
		t.Fatalf("first cancel failed: %v", err)
	}
	stockAfterFirst := s.products[1].Stock

	got, cancelled, err := s.CancelOrder(ctx, order.ID)
	if err != nil {
		t.Fatalf("second cancel failed: %v", err)
	}
	if cancelled {
		t.Fatalf("second cancelled = true, want false (already cancelled)")
	}
	if got.Status != StatusCancelled {
		t.Fatalf("status = %s, want cancelled", got.Status)
	}
	if s.products[1].Stock != stockAfterFirst {
		t.Fatalf("stock changed on second cancel: %d -> %d", stockAfterFirst, s.products[1].Stock)
	}
}

func TestCancelOrder_NotFound(t *testing.T) {
	s := newTestStore()
	_, _, err := s.CancelOrder(context.Background(), 999999)
	if err == nil {
		t.Fatalf("expected error for missing order")
	}
}
