package auth

import (
	"log"
	"net/http"

	"github.com/turos22/APIRESTFull_GoLang/internal/json"
)

type handler struct {
	service Service
}

func NewHandler(s Service) *handler {
	return &handler{service: s}
}

// GET
// me
func (h *handler) Me(w http.ResponseWriter, r *http.Request) {
	json.Write(w, http.StatusOK, nil)
}

// Post
// auth register
func (h *handler) Register(w http.ResponseWriter, r *http.Request) {
	var registro CreateUserParams
	if err := json.Read(r, &registro); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	usuarioCriado, err := h.service.Register(r.Context(), registro)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.Write(w, http.StatusOK, NewUserResposta(usuarioCriado))
}
// auth login
func (h *handler) Login(w http.ResponseWriter, r *http.Request) {
	var registro LoginUserParams
	if err := json.Read(r, &registro); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	token, err := h.service.Login(r.Context(), registro)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.Write(w, http.StatusOK, token)
}

// auth logout
func (h *handler) Logout(w http.ResponseWriter, r *http.Request) {
	var registro LoginUserParams
	if err := json.Read(r, &registro); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err := h.service.Logout(r.Context(), registro)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.Write(w, http.StatusOK, nil)
}

