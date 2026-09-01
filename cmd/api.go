package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/jwtauth/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	repo "github.com/turos22/APIRESTFull_GoLang/internal/adapters/postgresql/sqlc"
	"github.com/turos22/APIRESTFull_GoLang/internal/entidades/auth"
	"github.com/turos22/APIRESTFull_GoLang/internal/entidades/orders"
	"github.com/turos22/APIRESTFull_GoLang/internal/entidades/products"
)

func RequireRole(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, claims, err := jwtauth.FromContext(r.Context())
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			userRole, _ := claims["role"].(string)
			if userRole != role {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// mount - Metodos de chamada da api
func (app *application) mount() http.Handler {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.RequestID) //importante for rate limiting (passar requestID por todas funcoes)
	r.Use(middleware.RealIP)    // importante for rate limiting and analytics and tracing
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer) //recover from crashes
	r.Use(middleware.Timeout(60 * time.Second))

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
	}))
	//Services e Handlers
	ordersService := orders.NewService(repo.New(app.db), app.db)
	ordersHandler := orders.NewHandler(ordersService)

	productsService := products.NewService(repo.New(app.db))
	productHandler := products.NewHandler(productsService)

	authService := auth.NewService(repo.New(app.db), app.db)
	authHandler := auth.NewHandler(authService, app.jwt)

	//Rotas publicas
	//Gets
	r.Get("/products", productHandler.ListProducts)
	r.Get("/product/{id}", productHandler.FindProductById)
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("all good"))
	})

	//Post
	r.Post("/auth/register", authHandler.Register)
	r.Post("/auth/login", authHandler.Login)

	//Requisicoes com JWT
	r.Group(func(r chi.Router) {
		r.Use(jwtauth.Verifier(app.jwt))
		r.Use(jwtauth.Authenticator(app.jwt))

		//Gets
		r.Get("/auth/me", authHandler.Me)
		//Posts
		r.Post("/orders", ordersHandler.PlaceOrder)
		r.Post("/auth/logout", authHandler.Logout)

		r.Group(func(r chi.Router) {
			r.Use(RequireRole("vendedor"))

			r.Get("/me/products", productHandler.MeusProduct)
			r.Post("/products", productHandler.CreateProduct)
			r.Patch("/products", productHandler.UpdateProduct)
			r.Delete("/products", productHandler.DeleteProduct)

		})
	})

	return r
}

// run - Para rodar a api
func (app *application) run(h http.Handler) error {
	srv := &http.Server{
		Addr:         app.config.addr,
		Handler:      h,
		WriteTimeout: time.Second * 30,
		ReadTimeout:  time.Second * 10,
		IdleTimeout:  time.Second,
	}

	//Adicao de Signal para poder finalizar o servidor de forma correta
	// Criacao de notificacao para interromper
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	sair := make(chan os.Signal, 1)
	go func() {
		log.Printf("Server has started at addr %s", app.config.addr)
		if err := srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
		sair <- os.Interrupt
	}()

	//Espera alguma notificacaoi para desligar (1- nao subiu, 2- interrompeu)
	select {
	case <-sair:
		slog.Info("Servidor nao subiu")
	case <-ctx.Done():
		slog.Info("Servidor sera finalizado")
	}

	//cria contexto de 15s para terminar execucoes em background
	ctxDesligar, cancelar := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancelar()
	return srv.Shutdown(ctxDesligar)
}

type application struct {
	config config
	db     *pgxpool.Pool
	jwt    *jwtauth.JWTAuth
}

type config struct {
	addr string
	db   dbConfig
}

type dbConfig struct {
	dsn string //user, password, host, port, dbname
}
