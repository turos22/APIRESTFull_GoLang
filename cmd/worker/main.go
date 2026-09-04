package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/turos22/APIRESTFull_GoLang/internal/env"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	rdb := redis.NewClient(&redis.Options{Addr: env.GetString("GOOSE_REDIS", "localhost:6379")}) 

	err := rdb.XGroupCreateMkStream(ctx, "orders.created", "pagamentos", "0").Err()
	if err != nil && !strings.Contains(err.Error(), "BUSYGROUP") {
		panic(err)
	}

	consumerName := "worker-" + os.Getenv("HOSTNAME")

	for {
		select {
		case <-ctx.Done():
			return // SIGTERM chegou
		default:
		}

		streams, err := rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    "pagamentos",
			Consumer: consumerName,
			Streams:  []string{"orders.created", ">"},
			Count:    1,
			Block:    5 * time.Second,
		}).Result()

		if err == redis.Nil {
			continue 
		}
		if err != nil {
			log.Println(err)
			continue
		}

		for _, msg := range streams[0].Messages {
			log.Println(msg)

			//marcar que recebeu a chamada
			if err := rdb.XAck(ctx, "orders.created", "pagamentos", msg.ID).Err(); err != nil {
				log.Println("falha ao dar XAck:", err)
			}
		}
	}
}