package orders

import (
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	repo "github.com/turos22/APIRESTFull_GoLang/internal/adapters/postgresql/sqlc"
	JSON "github.com/turos22/APIRESTFull_GoLang/internal/json"
)

type handler struct {
	service Service
}

func NewHandler(s Service) *handler {
	return &handler{service: s}
}

func (h *handler) PlaceOrder(w http.ResponseWriter, r *http.Request) {
	var tempOrder createOrderParams
	if err := JSON.Read(r, &tempOrder); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	createdOrder, itensOrder, err := h.service.PlaceOrder(r.Context(), tempOrder)
	if err != nil {
		log.Println(err)

		if err == ErrProductNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	//2. return 201 created
	JSON.Write(w, http.StatusCreated, NovaOrdersResposta(createdOrder, itensOrder))
}

func (h *handler) OrderId(w http.ResponseWriter, r *http.Request){
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	order, err := h.service.GetOrderId(r.Context(), id)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	orderitens, err := h.service.GetItemsOrder(r.Context(), id)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	JSON.Write(w, http.StatusOK, NovaOrdersResposta(order, orderitens))
}

func (h *handler) OrdersMe(w http.ResponseWriter, r *http.Request){
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	orders, err := h.service.MeOrder(r.Context(), id)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var itemsTotal []repo.OrderItem
	for _, v := range orders {
		v, err := h.service.GetItemsOrder(r.Context(), v.ID)
		itemsTotal = append(itemsTotal, v...)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	JSON.Write(w, http.StatusOK, NovaListaOrdersResposta(orders, itemsTotal))
}