package orders

import (
	"context"

	repo "github.com/turos22/APIRESTFull_GoLang/internal/adapters/postgresql/sqlc"
)

type orderItem struct {
	ProductID int64 `json:"product_id"`
	Quantity  int32   `json:"quantity"`
}

type createOrderParams struct {
	CustomerID int64       `json:"customer_id"`
	Items      []orderItem `json:"items"`
}

type Service interface {
	PlaceOrder(ctx context.Context, temporder createOrderParams) (repo.Order, []repo.OrderItem, error)
	GetOrderId(ctx context.Context, id int64) (repo.Order, error)
	MeOrder(ctx context.Context, id int64) ([]repo.Order, error)
	GetItemsOrder(ctx context.Context, id int64) ([]repo.OrderItem, error)
}