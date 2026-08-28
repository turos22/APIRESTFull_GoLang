package products

import repo "github.com/turos22/APIRESTFull_GoLang/internal/adapters/postgresql/sqlc"

type ProductResposta struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Price       int64  `json:"price"`
	Quantity    int32  `json:"quantity"`
}


func NovaListaRespota (u []repo.Product) []ProductResposta {
	var items []ProductResposta
	for _, v := range u {
		items = append(items, NovaOrdersResposta(v))
	}
	return items
}


func NovaOrdersResposta(u repo.Product) ProductResposta {
	return ProductResposta{
		ID          : u.ID,
		Name        : u.Name,
		Description : u.Description.String,
		Price       : int64(u.PriceInCents),
		Quantity    : u.Quantity,
	}
}