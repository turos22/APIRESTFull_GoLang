package auth

import (
	"context"
	"fmt"

	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	repo "github.com/turos22/APIRESTFull_GoLang/internal/adapters/postgresql/sqlc"
)

type Service interface {
	Register(ctx context.Context, params CreateUserParams) (repo.User, error)
	Login(ctx context.Context, params LoginUserParams) (repo.User, error)
	Logout(ctx context.Context, params LoginUserParams) error
	Me(ctx context.Context, id int64) (repo.User, error)
}

type svc struct {
	repo repo.Queries
	db *pgxpool.Pool
}

func NewService(repo *repo.Queries, db *pgxpool.Pool) Service {
	return &svc{
		repo: *repo,
		db: db,
	}
}

func (s *svc) Register(ctx context.Context, params CreateUserParams) (repo.User, error) {

	if len(strings.TrimSpace(params.Name)) == 0 {
		return repo.User{}, fmt.Errorf("Name is required")
	}
	if len(strings.TrimSpace(params.Email)) == 0{
		return repo.User{}, fmt.Errorf("Email is required")
	}
	if len(strings.TrimSpace(params.Password)) == 0{
		return repo.User{}, fmt.Errorf("Senha is required")
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return repo.User{}, err
	}
	defer tx.Rollback(ctx)

	qtx := s.repo.WithTx(tx)
	
	Usuario, err := qtx.Register(ctx, repo.RegisterParams{
		Email: params.Email,
		Name:  params.Name,
		PasswordHash: params.Password,
		Role: params.Role,
	})
	if err != nil {
		return repo.User{}, err
	}
	tx.Commit(ctx)

	return Usuario, nil

}

func (s *svc) Login(ctx context.Context, params LoginUserParams) (repo.User, error) {
	return s.repo.FindUserByEmailPassword(ctx, repo.FindUserByEmailPasswordParams{
		Email: params.Email,
		PasswordHash: params.Password,
	})
}

func (s *svc) Logout(ctx context.Context, params LoginUserParams) error {
	usuario, err := s.repo.FindUserByEmailPassword(ctx, repo.FindUserByEmailPasswordParams{
		Email: params.Email,
		PasswordHash: params.Password,
	})

	if err != nil {
		return err
	}

	if usuario.ID == 0 {
		return fmt.Errorf("Usuário não encontrado")
	}

	return nil
}

func (s *svc) Me(ctx context.Context, id int64) (repo.User, error) {
	return s.repo.Me(ctx, id)
}
