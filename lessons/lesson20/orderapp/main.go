package main

import (
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gcron"
	"github.com/gogf/gf/v2/os/gtime"
)

type OrderStatus string

const (
	StatusPending   OrderStatus = "pending"
	StatusPaid      OrderStatus = "paid"
	StatusCancelled OrderStatus = "cancelled"
)

type Product struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Stock int    `json:"stock"`
}

type Order struct {
	ID             int64       `json:"id"`
	ProductID      int64       `json:"productId"`
	Quantity       int         `json:"quantity"`
	Status         OrderStatus `json:"status"`
	IdempotencyKey string      `json:"idempotencyKey"`
	ExpireAt       string      `json:"expireAt"`
}

type CreateOrderInput struct {
	ProductID      int64
	Quantity       int
	IdempotencyKey string
	ExpireAfter    time.Duration
}

type idempotencyRecord struct {
	OrderID int64
	Expires time.Time
}

type Store struct {
	mu          sync.Mutex
	nextOrderID int64
	products    map[int64]*Product
	orders      map[int64]*Order
	idem        map[string]idempotencyRecord
}

type apiResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

func main() {
	store := NewStore()

	if err := StartOrderJobs(context.Background(), store); err != nil {
		panic(err)
	}

	server := g.Server()
	server.SetPort(8010)
	server.Group("/api", func(group *ghttp.RouterGroup) {
		group.GET("/products/:id", handleGetProduct(store))
		group.GET("/orders/:id", handleGetOrder(store))
		group.POST("/orders", handleCreateOrder(store))
		group.POST("/orders/:id/pay", handlePayOrder(store))
		group.POST("/orders/:id/cancel", handleCancelOrder(store))
		group.POST("/jobs/cancel-expired", handleCancelExpired(store))
	})
	server.Run()
}

func NewStore() *Store {
	return &Store{
		nextOrderID: 1000,
		products: map[int64]*Product{
			1: {ID: 1, Name: "GoFrame Keyboard", Stock: 5},
		},
		orders: make(map[int64]*Order),
		idem:   make(map[string]idempotencyRecord),
	}
}

func StartOrderJobs(ctx context.Context, store *Store) error {
	entry, err := gcron.AddSingleton(
		ctx,
		"*/2 * * * * *",
		func(jobCtx context.Context) {
			cancelled, err := store.CancelExpired(jobCtx)
			if err != nil {
				g.Log().Error(jobCtx, "cancel expired orders failed", err)
				return
			}
			if cancelled > 0 {
				g.Log().Info(jobCtx, "expired orders cancelled", "count", cancelled)
			}
		},
		"cancel-expired-orders",
	)
	if err != nil {
		return gerror.Wrap(err, "register cancel expired orders job failed")
	}
	g.Log().Info(ctx, "order job registered", "name", entry.Name)
	return nil
}

func (s *Store) CreateOrder(ctx context.Context, input CreateOrderInput) (*Order, bool, error) {
	if input.IdempotencyKey == "" {
		return nil, false, gerror.New("Idempotency-Key header is required")
	}
	if input.ProductID <= 0 {
		return nil, false, gerror.New("productId must be greater than 0")
	}
	if input.Quantity <= 0 {
		return nil, false, gerror.New("quantity must be greater than 0")
	}
	if input.ExpireAfter <= 0 {
		input.ExpireAfter = 10 * time.Second
	}

	now := time.Now()

	s.mu.Lock()
	defer s.mu.Unlock()

	if record, ok := s.idem[input.IdempotencyKey]; ok && record.Expires.After(now) {
		order, ok := s.orders[record.OrderID]
		if !ok {
			return nil, false, gerror.New("idempotency record points to missing order")
		}
		return cloneOrder(order), true, nil
	}

	product, ok := s.products[input.ProductID]
	if !ok {
		return nil, false, gerror.New("product not found")
	}
	if product.Stock < input.Quantity {
		return nil, false, gerror.New("stock is not enough")
	}

	s.nextOrderID++
	order := &Order{
		ID:             s.nextOrderID,
		ProductID:      input.ProductID,
		Quantity:       input.Quantity,
		Status:         StatusPending,
		IdempotencyKey: input.IdempotencyKey,
		ExpireAt:       gtime.New(now.Add(input.ExpireAfter)).String(),
	}

	product.Stock -= input.Quantity
	s.orders[order.ID] = order
	s.idem[input.IdempotencyKey] = idempotencyRecord{
		OrderID: order.ID,
		Expires: now.Add(10 * time.Minute),
	}

	g.Log().Info(ctx, "order created", "orderId", order.ID, "stock", product.Stock)
	return cloneOrder(order), false, nil
}

func (s *Store) PayOrder(ctx context.Context, orderID int64) (*Order, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	order, ok := s.orders[orderID]
	if !ok {
		return nil, gerror.New("order not found")
	}
	if order.Status == StatusCancelled {
		return nil, gerror.New("cancelled order cannot be paid")
	}
	if order.Status == StatusPaid {
		return cloneOrder(order), nil
	}
	order.Status = StatusPaid
	g.Log().Info(ctx, "order paid", "orderId", order.ID)
	return cloneOrder(order), nil
}

func (s *Store) CancelExpired(ctx context.Context) (int, error) {
	now := time.Now()
	expiredIDs := make([]int64, 0)

	s.mu.Lock()
	for _, order := range s.orders {
		expireAt, err := time.ParseInLocation("2006-01-02 15:04:05", order.ExpireAt, time.Local)
		if err != nil {
			s.mu.Unlock()
			return 0, gerror.Wrap(err, "parse order expireAt failed")
		}
		if order.Status == StatusPending && !expireAt.After(now) {
			expiredIDs = append(expiredIDs, order.ID)
		}
	}
	s.mu.Unlock()

	cancelled := 0
	for _, orderID := range expiredIDs {
		ok, err := s.cancelOne(ctx, orderID)
		if err != nil {
			g.Log().Error(ctx, "cancel one expired order failed", "orderId", orderID, err)
			continue
		}
		if ok {
			cancelled++
		}
	}
	return cancelled, nil
}

func (s *Store) cancelOne(ctx context.Context, orderID int64) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	order, ok := s.orders[orderID]
	if !ok {
		return false, gerror.New("order not found")
	}

	expireAt, err := time.ParseInLocation("2006-01-02 15:04:05", order.ExpireAt, time.Local)
	if err != nil {
		return false, gerror.Wrap(err, "parse order expireAt failed")
	}
	if order.Status != StatusPending || expireAt.After(time.Now()) {
		return false, nil
	}

	product, ok := s.products[order.ProductID]
	if !ok {
		return false, gerror.New("product not found")
	}

	order.Status = StatusCancelled
	product.Stock += order.Quantity
	g.Log().Info(ctx, "order cancelled and stock returned", "orderId", order.ID, "stock", product.Stock)
	return true, nil
}

func (s *Store) CancelOrder(ctx context.Context, orderID int64) (*Order, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	order, ok := s.orders[orderID]
	if !ok {
		return nil, false, gerror.New("order not found")
	}

	switch order.Status {
	case StatusCancelled:
		return cloneOrder(order), false, nil
	case StatusPaid:
		return nil, false, gerror.New("paid order cannot be cancelled")
	}

	product, ok := s.products[order.ProductID]
	if !ok {
		return nil, false, gerror.New("product not found")
	}

	order.Status = StatusCancelled
	product.Stock += order.Quantity
	g.Log().Info(ctx, "order cancelled by user", "orderId", order.ID, "stock", product.Stock)
	return cloneOrder(order), true, nil
}

func handleGetProduct(store *Store) func(r *ghttp.Request) {
	return func(r *ghttp.Request) {
		productID := r.GetRouter("id").Int64()

		store.mu.Lock()
		defer store.mu.Unlock()

		product, ok := store.products[productID]
		if !ok {
			writeJSON(r, 51, "product not found", nil)
			return
		}
		copyProduct := *product
		writeJSON(r, 0, "ok", copyProduct)
	}
}

func handleCreateOrder(store *Store) func(r *ghttp.Request) {
	return func(r *ghttp.Request) {
		expireSeconds := r.Get("expireSeconds", 10).Int()
		order, repeated, err := store.CreateOrder(r.Context(), CreateOrderInput{
			ProductID:      r.Get("productId", 1).Int64(),
			Quantity:       r.Get("quantity", 1).Int(),
			IdempotencyKey: r.Header.Get("Idempotency-Key"),
			ExpireAfter:    time.Duration(expireSeconds) * time.Second,
		})
		if err != nil {
			writeJSON(r, 52, err.Error(), nil)
			return
		}
		writeJSON(r, 0, "ok", g.Map{
			"repeated": repeated,
			"order":    order,
		})
	}
}

func handleGetOrder(store *Store) func(r *ghttp.Request) {
	return func(r *ghttp.Request) {
		orderID := r.GetRouter("id").Int64()

		store.mu.Lock()
		defer store.mu.Unlock()

		order, ok := store.orders[orderID]
		if !ok {
			writeJSON(r, 51, "order not found", nil)
			return
		}
		writeJSON(r, 0, "ok", cloneOrder(order))
	}
}

func handlePayOrder(store *Store) func(r *ghttp.Request) {
	return func(r *ghttp.Request) {
		orderID, err := strconv.ParseInt(r.GetRouter("id").String(), 10, 64)
		if err != nil {
			writeJSON(r, 52, "invalid order id", nil)
			return
		}

		order, err := store.PayOrder(r.Context(), orderID)
		if err != nil {
			writeJSON(r, 52, err.Error(), nil)
			return
		}
		writeJSON(r, 0, "ok", order)
	}
}

func handleCancelOrder(store *Store) func(r *ghttp.Request) {
	return func(r *ghttp.Request) {
		orderID, err := strconv.ParseInt(r.GetRouter("id").String(), 10, 64)
		if err != nil {
			writeJSON(r, 52, "invalid order id", nil)
			return
		}

		order, cancelled, err := store.CancelOrder(r.Context(), orderID)
		if err != nil {
			writeJSON(r, 52, err.Error(), nil)
			return
		}
		writeJSON(r, 0, "ok", g.Map{
			"cancelled": cancelled,
			"order":     order,
		})
	}
}

func handleCancelExpired(store *Store) func(r *ghttp.Request) {
	return func(r *ghttp.Request) {
		cancelled, err := store.CancelExpired(r.Context())
		if err != nil {
			writeJSON(r, 52, err.Error(), nil)
			return
		}
		writeJSON(r, 0, "ok", g.Map{"cancelled": cancelled})
	}
}

func cloneOrder(order *Order) *Order {
	copyOrder := *order
	return &copyOrder
}

func writeJSON(r *ghttp.Request, code int, message string, data any) {
	r.Response.WriteJson(apiResponse{
		Code:    code,
		Message: message,
		Data:    data,
	})
}
