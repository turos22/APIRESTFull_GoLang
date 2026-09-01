package products

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	repo "github.com/turos22/APIRESTFull_GoLang/internal/adapters/postgresql/sqlc"
)

type Service interface {
	ListProducts(ctx context.Context, params repo.ListProductsParams) ([]repo.ListProductsRow, error)
	FindProductByID(ctx context.Context, id int64) (repo.Product, error)
	CreateProduct(ctx context.Context, params repo.CreateProdutoParams) (repo.Product, error)
	UpdateProduct(ctx context.Context, params repo.UpdateProductParams) (repo.Product, error)
	DeleteProduct(ctx context.Context, id int64) error
	MeProduct(ctx context.Context, id int64) ([]repo.Product, error)
}

type svc struct {
	repo repo.Querier
}

func NewService(repo repo.Querier) Service {
	return &svc{repo: repo}
}

func (s *svc) ListProducts(ctx context.Context, params repo.ListProductsParams) ([]repo.ListProductsRow, error) {
	return s.repo.ListProducts(ctx, repo.ListProductsParams{})
}

func (s *svc) FindProductByID(ctx context.Context, id int64) (repo.Product, error) {
	return s.repo.FindProductByID(ctx, id)
}

func (s *svc) CreateProduct(ctx context.Context, params repo.CreateProdutoParams) (repo.Product, error) {
	return s.repo.CreateProduto(ctx, params)
}

func (s *svc) UpdateProduct(ctx context.Context, params repo.UpdateProductParams) (repo.Product, error) {
	return s.repo.UpdateProduct(ctx, params)
}

func (s *svc) DeleteProduct(ctx context.Context, id int64) error {
	return s.repo.DeleteProduct(ctx, id)
}

func (s *svc) MeProduct(ctx context.Context, id int64) ([]repo.Product, error) {
	return s.repo.Meproducts(ctx, pgtype.Int8{Int64: id})
}

