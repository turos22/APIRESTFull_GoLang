package auth

import (
	"github.com/jackc/pgx/v5/pgtype"
	repo "github.com/turos22/APIRESTFull_GoLang/internal/adapters/postgresql/sqlc"
)
type UserResposta struct {
	Id        int64              `json:"id"`
	Name      string             `json:"name"`
	Email     string             `json:"email"`
	Role      string             `json:"role"`
	CreatedAt pgtype.Timestamptz `json:"createdAt"` 
}

func NewUserResposta(u repo.User) UserResposta {
	return UserResposta{
		Id:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		Role:      u.Role,
		CreatedAt: u.CreatedAt, 
	}
}