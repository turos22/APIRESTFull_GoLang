package orders

import repo "github.com/turos22/APIRESTFull_GoLang/internal/adapters/postgresql/sqlc"

type OrdersResposta struct {
	OrderID    int64               `json:"order_id"`
	CustomerID int64               `json:"customer_id"`
	Items      []orderItemResposta `json:"items"`
}

type orderItemResposta struct {
	ProductID int64 `json:"product_id"`
	Quantity  int32 `json:"quantity"`
}

func NovaOrderItemResposta(u repo.OrderItem) orderItemResposta {
	return orderItemResposta{
		ProductID: u.ProductID,
		Quantity:  u.Quantity,
	}
}

func NovaOrdersResposta(u repo.Order, i []repo.OrderItem) OrdersResposta {
	var items []orderItemResposta
	for _, v := range i {
		items = append(items, NovaOrderItemResposta(v))
	}
	return OrdersResposta{
		OrderID:    u.ID,
		CustomerID: u.CustomerID,
		Items:      items,
	}
}