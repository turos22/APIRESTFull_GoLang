package orders

import (
	"log"
	"net/http"

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

