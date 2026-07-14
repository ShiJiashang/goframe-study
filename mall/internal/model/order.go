package model

type OrderCreateInput struct {
	UserID    int64
	ProductID int64
	Quantity  int
}

type OrderCreateOutput struct {
	OrderID   int64
	OrderNo   string
	TotalCent int64
}
