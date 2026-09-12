package orders

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/redis/go-redis/v9"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	repo "github.com/turos22/APIRESTFull_GoLang/internal/adapters/postgresql/sqlc"
)

const (
	estoqueInicial = 10
	tentativas     = 50
)

func TestVendaConcorrenteNaoFuraEstoque(t *testing.T) {
	ctx := context.Background()

	pool := bancoDeTeste(t, ctx)
	q := repo.New(pool)

	comprador := criarComprador(t, ctx, q)
	produto := criarProduto(t, ctx, q, estoqueInicial)

	svc := NewService(q, pool, redis.NewClient(&redis.Options{Addr: "127.0.0.1:1"}))

	var sucessos, semEstoque, outrosErros int64

	var wg sync.WaitGroup
	largada := make(chan struct{})

	for i := 0; i < tentativas; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()
			<-largada

			_, _, err := svc.PlaceOrder(ctx, createOrderParams{
				CustomerID: comprador.ID,
				Items:      []orderItem{{ProductID: produto.ID, Quantity: 1}},
			})
			switch {
			case err == nil:
				atomic.AddInt64(&sucessos, 1)
			case errors.Is(err, ErrProductOutOfStock):
				atomic.AddInt64(&semEstoque, 1)
			default:
				t.Logf("erro inesperado: %v", err)
				atomic.AddInt64(&outrosErros, 1)
			}
		}()
	}

	close(largada)
	wg.Wait()

	var estoqueFinal int32
	err := pool.QueryRow(ctx,
		`SELECT quantity FROM products WHERE id = $1`, produto.ID).Scan(&estoqueFinal)
	if err != nil {
		t.Fatalf("ler o estoque final: %v", err)
	}

	var unidadesVendidas int64
	err = pool.QueryRow(ctx,
		`SELECT coalesce(sum(quantity), 0) FROM order_items WHERE product_id = $1`,
		produto.ID).Scan(&unidadesVendidas)
	if err != nil {
		t.Fatalf("somar as unidades vendidas: %v", err)
	}

	var linhasDeItem int64
	err = pool.QueryRow(ctx,
		`SELECT count(*) FROM order_items WHERE product_id = $1`,
		produto.ID).Scan(&linhasDeItem)
	if err != nil {
		t.Fatalf("contar order_items: %v", err)
	}

	t.Logf("sucessos=%d  semEstoque=%d  outrosErros=%d  estoqueFinal=%d  vendidas=%d",
		sucessos, semEstoque, outrosErros, estoqueFinal, unidadesVendidas)

	baixaNoEstoque := int64(estoqueInicial) - int64(estoqueFinal)

	if baixaNoEstoque != unidadesVendidas {
		t.Errorf("a conta nao fecha: o estoque caiu %d unidade(s), mas foram vendidas %d",
			baixaNoEstoque, unidadesVendidas)
	}

	if sucessos != estoqueInicial {
		t.Errorf("pedidos aceitos: esperava %d, veio %d", estoqueInicial, sucessos)
	}

	if semEstoque != tentativas-estoqueInicial {
		t.Errorf("recusas por falta de estoque: esperava %d, veio %d",
			tentativas-estoqueInicial, semEstoque)
	}
	if outrosErros != 0 {
		t.Errorf("houve %d erro(s) inesperado(s) — veja os t.Log acima", outrosErros)
	}

	if estoqueFinal != 0 {
		t.Errorf("estoque final: esperava 0, veio %d", estoqueFinal)
	}
	if estoqueFinal < 0 {
		t.Errorf("o estoque ficou NEGATIVO: %d", estoqueFinal)
	}

	if linhasDeItem != int64(estoqueInicial) {
		t.Errorf("linhas em order_items: esperava %d, veio %d", estoqueInicial, linhasDeItem)
	}
}

func bancoDeTeste(t *testing.T, ctx context.Context) *pgxpool.Pool {
	t.Helper()

	container, err := postgres.Run(ctx,
		"postgres:17-alpine",
		postgres.WithDatabase("api_ecom_test"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		postgres.BasicWaitStrategies(),
	)
	if err != nil {
		t.Fatalf("subir o postgres de teste: %v", err)
	}

	t.Cleanup(func() {
		if err := testcontainers.TerminateContainer(container); err != nil {
			t.Logf("falha ao derrubar o container: %v", err)
		}
	})

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("montar o dsn: %v", err)
	}

	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Fatalf("interpretar o dsn: %v", err)
	}

	cfg.MaxConns = 25

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatalf("abrir o pool: %v", err)
	}
	t.Cleanup(pool.Close)

	migrar(t, ctx, pool)

	return pool
}

func migrar(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()

	db := stdlib.OpenDBFromPool(pool)
	defer db.Close()

	goose.SetLogger(goose.NopLogger())

	if err := goose.SetDialect("postgres"); err != nil {
		t.Fatalf("dialeto do goose: %v", err)
	}

	if err := goose.UpContext(ctx, db, "../../adapters/postgresql/migrations"); err != nil {
		t.Fatalf("aplicar as migrations: %v", err)
	}
}

func criarComprador(t *testing.T, ctx context.Context, q *repo.Queries) repo.User {
	t.Helper()

	usuario, err := q.Register(ctx, repo.RegisterParams{
		Email:        "comprador@teste.com",
		PasswordHash: "hash-de-teste",
		Name:         "Comprador",
		Role:         "comprador",
	})
	if err != nil {
		t.Fatalf("criar comprador: %v", err)
	}
	return usuario
}

func criarProduto(t *testing.T, ctx context.Context, q *repo.Queries, estoque int32) repo.Product {
	t.Helper()

	produto, err := q.CreateProduto(ctx, repo.CreateProdutoParams{
		Name:         "Camisa",
		PriceInCents: 5000,
		Quantity:     estoque,
		Description:  pgtype.Text{String: "produto do teste de concorrencia", Valid: true},
		ImageUrl:     pgtype.Text{String: "http://exemplo/camisa.png", Valid: true},
		CategoryID:   pgtype.Int8{Int64: 1, Valid: true},
		Active:       pgtype.Bool{Bool: true, Valid: true},
		SellerID:     pgtype.Int8{},
	})
	if err != nil {
		t.Fatalf("criar produto: %v", err)
	}

	return produto
}
