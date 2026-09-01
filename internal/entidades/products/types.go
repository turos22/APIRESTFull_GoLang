package products


type produtoparams struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	PriceInCents int    `json:"price_in_cents"`
	Quantity    int    `json:"quantity"`
	Description string `json:"description"`
	ImageURL    string `json:"image_url"`
	CategoryID  int64  `json:"category_id"`
	Active      bool   `json:"active"`
	SellerID    int64  `json:"seller_id"`
}
