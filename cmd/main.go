package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/go-chi/jwtauth/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	//"github.com/jackc/pgx/v5/pgxpool"
	"github.com/turos22/APIRESTFull_GoLang/internal/env"
)

func main() {
	ctx := context.Background()

	cfg := config{
		addr: ":8080",
		db: dbConfig{
			dsn: env.GetString("GOOSE_DBSTRING", "postgres://postgres:postgres@localhost:5434/api_ecom?sslmode=disable"),
		},
		redisAddr: env.GetString("GOOSE_REDIS", "localhost:6379"),
	}

	poolcfg, err := pgxpool.ParseConfig(cfg.db.dsn)
	if err != nil { panic(err) }

	//pool de conexoes para evitar abrir varias conexoes por requisicao, limitando
	poolcfg.MaxConns        =  int32(env.GetInt("MAX_CONNS", 10))
	poolcfg.MinConns        =  int32(env.GetInt("MIN_CONNS", 2))
	poolcfg.MaxConnLifetime = time.Duration(env.GetInt("MAXCONNLIFE", 30)) * time.Minute
	poolcfg.MaxConnIdleTime = time.Duration(env.GetInt("MAXCONIDLE", 5)) * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, poolcfg)
	if err != nil { panic(err) }
	defer pool.Close()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault((logger))


	logger.Info("Connected to database", "dsn", cfg.db.dsn)
	jwtauth := jwtauth.New("HS256", []byte("secret"), nil)

	rdb := redis.NewClient(&redis.Options{
		Addr: cfg.redisAddr, // vem de env, ex: "redis:6379"
	})

	api := application{
		config: cfg,
		db: pool,
		jwt: jwtauth,
		rdb : rdb,
	}
	//montar o servidor
	//Subir o servidor
	if err := api.run(api.mount()); err != nil {
		slog.Error("Server failed to start", "error", err)
		os.Exit(1)
	}

}
