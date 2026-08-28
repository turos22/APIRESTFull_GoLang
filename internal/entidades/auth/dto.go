package auth

import repo "github.com/turos22/APIRESTFull_GoLang/internal/adapters/postgresql/sqlc"

type UserResposta struct {
	Id    int64  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func NewUserResposta(u repo.User) UserResposta {
	return UserResposta{
		Id:    u.ID,
		Name:  u.Name,
		Email: u.Email,
	}
}