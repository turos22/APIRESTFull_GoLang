package products

import (
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
	repo "github.com/turos22/APIRESTFull_GoLang/internal/adapters/postgresql/sqlc"
	"github.com/turos22/APIRESTFull_GoLang/internal/json"
)

type handler struct {
	service Service
}

func NewHandler(s Service) *handler {
	return &handler{service: s}
}

func toPgText(s string) pgtype.Text {
	return pgtype.Text{String: s, Valid: s != ""}
}

func toPgInt8(v int64) pgtype.Int8 {
	return pgtype.Int8{Int64: v, Valid: v != 0}
}

func toPgInt4(v int32) pgtype.Int4 {
	return pgtype.Int4{Int32: v, Valid: v != 0}
}

func parseIntOrDefault(s string, def int) int {
	n, err := strconv.Atoi(s)
	if err != nil || n <= 0 {
		return def
	}
	return n
}

func VerificarProdut(produto produtoparams) string {
	
	if produto.Name == "" {
		return "name is required"
	}

	if produto.Description == "" {
		return "description is required"
	}

	if produto.PriceInCents == 0 {
		return "price_in_cents is required"
	}

	if produto.Quantity  == 0{
		return "quantity is required"
	}

	if produto.CategoryID == 0 {
		return "category_id is required"
	}
	
	return ""
}


func (h *handler) ListProducts(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	nomePesquisa := query.Get("name")

	var categoriaID int64
	if v := query.Get("categoriaid"); v != "" {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			http.Error(w, "categoriaid inválido", http.StatusBadRequest)
			return
		}
		categoriaID = id
	}

	var minPriceCents, maxPriceCents int32
	if v := query.Get("minprice"); v != "" {
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			http.Error(w, "minprice inválido", http.StatusBadRequest)
			return
		}
		minPriceCents = int32(f * 100)
	}
	if v := query.Get("maxprice"); v != "" {
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			http.Error(w, "maxprice inválido", http.StatusBadRequest)
			return
		}
		maxPriceCents = int32(f * 100)
	}

	limit := parseIntOrDefault(query.Get("limit"), 20)
	if limit > 100 {
		limit = 100
	}
	page := parseIntOrDefault(query.Get("page"), 1)
	offset := (page - 1) * limit

	params := repo.ListProductsParams{
		Search:     toPgText(nomePesquisa),
		CategoryID: toPgInt8(categoriaID),
		MinPrice:   toPgInt4(minPriceCents),
		MaxPrice:   toPgInt4(maxPriceCents),
		Limit:      int32(limit),
		Offset:     int32(offset),
	}

	products, err := h.service.ListProducts(r.Context(), params)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.Write(w, http.StatusOK, NovaListaRespota(products))
}


func (h *handler) FindProductById(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)

	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	product, err := h.service.FindProductByID(r.Context(), id)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.Write(w, http.StatusOK, NewProductResposta(product))
}

func (h *handler) CreateProduct(w http.ResponseWriter, r *http.Request){
	var produto produtoparams

	if err := json.Read(r, &produto); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	erro := VerificarProdut(produto)
	if erro != "" {
		http.Error(w, erro, http.StatusBadRequest)
		return
	}


	createdProduct, err := h.service.CreateProduct(r.Context(), repo.CreateProdutoParams{
		Name:           produto.Name,
		Description:    toPgText(produto.Description),
		PriceInCents:   int32(produto.PriceInCents),
		Quantity:       int32(produto.Quantity),
		CategoryID:     toPgInt8(produto.CategoryID),
		ImageUrl:       toPgText(produto.ImageURL),
		SellerID:       toPgInt8(produto.SellerID),
		Active:         pgtype.Bool{Bool: produto.Active},
	})
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.Write(w, http.StatusCreated, NewProductResposta(createdProduct))

}

func (h *handler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
		var produto produtoparams

	if err := json.Read(r, &produto); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	erro := VerificarProdut(produto)
	if erro != "" {
		http.Error(w, erro, http.StatusBadRequest)
		return
	}


	updatedProduct, err := h.service.UpdateProduct(r.Context(), repo.UpdateProductParams{
		Name:           produto.Name,
		Description:    toPgText(produto.Description),
		PriceInCents:   int32(produto.PriceInCents),
		Quantity:       int32(produto.Quantity),
		CategoryID:     toPgInt8(produto.CategoryID),
		ImageUrl:       toPgText(produto.ImageURL),
		SellerID:       toPgInt8(produto.SellerID),
		Active:         pgtype.Bool{Bool: produto.Active},
		ID: produto.ID,
	})
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.Write(w, http.StatusCreated, NewProductResposta(updatedProduct))
}

func (h *handler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)

	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = h.service.DeleteProduct(r.Context(), id)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.Write(w, http.StatusOK, nil)
}

func (h *handler) MeusProduct(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)

	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	products, err := h.service.MeProduct(r.Context(), id)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.Write(w, http.StatusOK, NewListProductsRespostas(products))
}