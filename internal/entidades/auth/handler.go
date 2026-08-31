package auth

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	"github.com/turos22/APIRESTFull_GoLang/internal/json"
)

type handler struct {
	service Service
	jwt *jwtauth.JWTAuth
}


func NewHandler(s Service, jwt *jwtauth.JWTAuth) *handler {
	return &handler{
		service: s,
		jwt: jwt,
	}
}

func (h *handler) DevolverCookieJWT(w http.ResponseWriter, name string, userid int64, role string) {
	const duracao = 24 * time.Hour
	expira := time.Now().Add(duracao)

	_, tokenString, err := h.jwt.Encode(map[string]interface{}{
		"sub":  userid,
		"role": role,
		"exp":  expira.Unix(),
	})
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "jwt",
		Value:    tokenString,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   false, 
		MaxAge:   int(duracao.Seconds()),
		Expires:  expira,
	})
}

// GET
// me
func (h *handler) Me(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	//Deve pegar o ID do Token na verdade

	usuario, err := h.service.Me(r.Context(), id)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.Write(w, http.StatusOK, NewUserResposta(usuario))
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

	h.DevolverCookieJWT(w, usuarioCriado.Name, usuarioCriado.ID, usuarioCriado.Role)

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

	usuario, err := h.service.Login(r.Context(), registro)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.DevolverCookieJWT(w, usuario.Name, usuario.ID, usuario.Role)
   
	json.Write(w, http.StatusOK, nil)
}

// auth logout
func (h *handler) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "jwt",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   false,
		MaxAge:   -1, // expira imediatamente
	})
	w.WriteHeader(http.StatusNoContent)
}

