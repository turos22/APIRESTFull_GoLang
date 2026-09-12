package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand/v2"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	repo "github.com/turos22/APIRESTFull_GoLang/internal/adapters/postgresql/sqlc"
	"github.com/turos22/APIRESTFull_GoLang/internal/env"
)

const (
	streamDePedidos = "orders.created"
	streamDeMortos  = "orders.dead"
	grupo           = "pagamentos"

	maxTentativas = 3
	chanceDeFalha = 0.10
)

var errPagamentoRecusado = errors.New("pagamento recusado apos todas as tentativas")

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	rdb := redis.NewClient(&redis.Options{Addr: env.GetString("GOOSE_REDIS", "localhost:6379")})

	dsn := env.GetString("GOOSE_DBSTRING", "postgres://postgres:postgres@localhost:5434/api_ecom?sslmode=disable")
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		panic(err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		panic(err)
	}

	q := repo.New(pool)

	err = rdb.XGroupCreateMkStream(ctx, streamDePedidos, grupo, "0").Err()
	if err != nil && !strings.Contains(err.Error(), "BUSYGROUP") {
		panic(err)
	}

	consumerName := "worker-" + os.Getenv("HOSTNAME")
	log.Printf("worker %s ouvindo %s", consumerName, streamDePedidos)

	for {
		select {
		case <-ctx.Done():
			return // SIGTERM chegou
		default:
		}

		streams, err := rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    grupo,
			Consumer: consumerName,
			Streams:  []string{streamDePedidos, ">"},
			Count:    1,
			Block:    5 * time.Second,
		}).Result()

		if err == redis.Nil {
			continue
		}
		if err != nil {
			if ctx.Err() != nil {
				return // o contexto caiu junto com o sinal
			}
			log.Println(err)
			continue
		}

		for _, msg := range streams[0].Messages {
			tratar(ctx, rdb, q, msg)
		}
	}
}

func tratar(ctx context.Context, rdb *redis.Client, q *repo.Queries, msg redis.XMessage) {
	pedidoID, err := idDoPedido(msg)
	if err != nil {
		log.Printf("mensagem %s descartada: %v", msg.ID, err)
		confirmar(ctx, rdb, msg.ID)
		return
	}

	err = processarPedido(ctx, q, pedidoID)

	if errors.Is(err, errPagamentoRecusado) {
		erroDaStream := rdb.XAdd(ctx, &redis.XAddArgs{
			Stream: streamDeMortos,
			Values: map[string]interface{}{
				"order_id": pedidoID,
				"motivo":   err.Error(),
			},
		}).Err()
		if erroDaStream != nil {
			log.Printf("pedido %d: falha ao publicar em %s: %v", pedidoID, streamDeMortos, erroDaStream)
		}
		confirmar(ctx, rdb, msg.ID)
		return
	}

	if err != nil {

		log.Printf("pedido %d: %v", pedidoID, err)
		return
	}

	confirmar(ctx, rdb, msg.ID)
}

func processarPedido(ctx context.Context, q *repo.Queries, pedidoID int64) error {
	if err := cobrar(pedidoID); err != nil {
		if erroDeGravacao := gravarStatus(ctx, q, pedidoID, "falhou"); erroDeGravacao != nil {
			return erroDeGravacao
		}
		return err
	}

	esteira := []struct {
		status string
		espera time.Duration
	}{
		{status: "pago", espera: 0},
		{status: "separando", espera: 2 * time.Second},
		{status: "enviado", espera: 2 * time.Second},
	}

	for _, etapa := range esteira {
		time.Sleep(etapa.espera)
		if err := gravarStatus(ctx, q, pedidoID, etapa.status); err != nil {
			return err
		}
	}

	return nil
}

func cobrar(pedidoID int64) error {
	for tentativa := 1; tentativa <= maxTentativas; tentativa++ {
		time.Sleep(time.Duration(2000+rand.IntN(3001)) * time.Millisecond)

		if rand.Float64() >= chanceDeFalha {
			return nil
		}

		log.Printf("pedido %d: cobranca recusada na tentativa %d de %d",
			pedidoID, tentativa, maxTentativas)
	}

	return errPagamentoRecusado
}

func gravarStatus(ctx context.Context, q *repo.Queries, pedidoID int64, status string) error {
	err := q.AtualizarStatusDoPedido(ctx, repo.AtualizarStatusDoPedidoParams{
		Status: pgtype.Text{String: status, Valid: true},
		ID:     pedidoID,
	})
	if err != nil {
		return fmt.Errorf("gravar status %q do pedido %d: %w", status, pedidoID, err)
	}

	log.Printf("pedido %d: status agora e %q", pedidoID, status)
	return nil
}

func idDoPedido(msg redis.XMessage) (int64, error) {
	cru, existe := msg.Values["order_id"]
	if !existe {
		return 0, fmt.Errorf("mensagem sem order_id")
	}
	return strconv.ParseInt(fmt.Sprint(cru), 10, 64)
}

func confirmar(ctx context.Context, rdb *redis.Client, msgID string) {
	if err := rdb.XAck(ctx, streamDePedidos, grupo, msgID).Err(); err != nil {
		log.Println("falha ao dar XAck:", err)
	}
}
