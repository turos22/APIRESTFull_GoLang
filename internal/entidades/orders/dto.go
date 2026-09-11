package orders

import (
	"time"

	repo "github.com/turos22/APIRESTFull_GoLang/internal/adapters/postgresql/sqlc"
)

type OrdersResposta struct {
	OrderID    int64               `json:"order_id"`
	CustomerID int64               `json:"customer_id"`
	Status     string              `json:"status"`
	TotalCents int32               `json:"total_cents"`
	CreatedAt  *time.Time          `json:"created_at"`
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
	// Slice nao-nil para a resposta trazer [] e nao null: o front faz .map
	// direto em items.
	items := make([]orderItemResposta, 0)
	for _, v := range i {
		if v.OrderID != u.ID {
			continue
		}
		items = append(items, NovaOrderItemResposta(v))
	}

	// pgtype.Timestamptz serializaria como objeto ({"Time":…,"Valid":…}).
	// O front espera string RFC3339, que e o que *time.Time entrega.
	var criadoEm *time.Time
	if u.CreatedAt.Valid {
		t := u.CreatedAt.Time
		criadoEm = &t
	}

	return OrdersResposta{
		OrderID:    u.ID,
		CustomerID: u.CustomerID,
		Status:     u.Status.String, // NULL vira "", e o front cai no fallback 'pendente'
		TotalCents: u.TotalCents.Int32,
		CreatedAt:  criadoEm,
		Items:      items,
	}
}

func NovaListaOrdersResposta(u []repo.Order, i []repo.OrderItem) []OrdersResposta {
	items := make([]OrdersResposta, 0, len(u))
	for _, v := range u {
		items = append(items, NovaOrdersResposta(v, i))
	}
	return items
}
