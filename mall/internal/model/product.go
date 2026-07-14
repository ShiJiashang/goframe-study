package model

type ProductListInput struct {
	Page       int
	Size       int
	Name       string
	CategoryID int64
	MinPrice   int64
	MaxPrice   int64
}

type ProductListItem struct {
	ID         int64
	CategoryID int64
	Name       string
	PriceCent  int64
	Stock      int
	Status     int
}

type ProductListOutput struct {
	List  []ProductListItem
	Total int
}

type ProductCreateInput struct {
	CategoryID int64
	Name       string
	PriceCent  int64
	Stock      int
}

type ProductCreateOutput struct {
	ID int64
}

type ProductDetailInput struct {
	ID int64
}

type ProductDetailOutput struct {
	ID         int64
	CategoryID int64
	Name       string
	PriceCent  int64
	Stock      int
	Status     int
}

type ProductUpdateInput struct {
	ID         int64
	CategoryID int64
	Name       string
	PriceCent  int64
	Stock      int
	Status     int
}

type ProductUpdateOutput struct{}

type ProductDeleteInput struct {
	ID int64
}

type ProductDeleteOutput struct{}
