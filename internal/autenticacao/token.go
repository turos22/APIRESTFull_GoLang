package autenticacao

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/jwtauth/v5"
)


func IDDoUsuario(r *http.Request) (int64, error) {
	_, claims, err := jwtauth.FromContext(r.Context())
	if err != nil {
		return 0, err
	}

	sub, ok := claims["sub"].(string)
	if !ok {
		return 0, fmt.Errorf("claim sub ausente ou em formato invalido")
	}

	return strconv.ParseInt(sub, 10, 64)
}

func RoleDoUsuario(r *http.Request) (string, error) {
	_, claims, err := jwtauth.FromContext(r.Context())
	if err != nil {
		return "", err
	}

	role, ok := claims["role"].(string)
	if !ok {
		return "", fmt.Errorf("claim role ausente ou em formato invalido")
	}

	return role, nil
}
