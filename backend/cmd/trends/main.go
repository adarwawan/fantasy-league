// Command trends is the isolated FPL transfer-velocity service. It runs its own
// HTTP server + poller on a separate port (default 8081), sharing the Postgres
// pool and Redis cache with the main app but owning only the trends_* tables.
package main

import (
	"context"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-co-op/gocron/v2"

	"fantasy-league/internal/config"
	"fantasy-league/internal/store"
	"fantasy-league/internal/trends"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()

	pg, err := store.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("postgres: %v", err)
	}
	defer pg.Close()

	cache, err := store.NewCache(cfg.RedisURL)
	if err != nil {
		log.Fatalf("redis: %v", err)
	}

	tStore := trends.NewStore(pg.Pool())
	client := trends.NewClient()
	poller := trends.NewPoller(client, tStore)
	handler := trends.NewHandler(tStore, cache, client, poller, cfg.TrendsSyncSecret)

	// Poller — fires every TrendsPollInterval; no-ops unless a session is active.
	scheduler, err := gocron.NewScheduler()
	if err != nil {
		log.Fatalf("scheduler: %v", err)
	}
	if _, err := scheduler.NewJob(
		gocron.DurationJob(cfg.TrendsPollInterval),
		gocron.NewTask(func() { poller.Tick(context.Background()) }),
	); err != nil {
		log.Fatalf("schedule poll job: %v", err)
	}
	scheduler.Start()

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.CORSAllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Content-Type", "X-Sync-Secret"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	r.Route("/api/trends", func(r chi.Router) {
		r.Get("/session", handler.Session)
		r.Get("/leaders", handler.Leaders)
		r.Get("/player/{extId}/series", handler.Series)
		r.Post("/session", handler.Arm) // secret-guarded in the handler
		r.Post("/poll", handler.Poll)   // secret-guarded; manual capture
	})

	log.Printf("trends service listening on :%s", cfg.TrendsPort)
	if err := http.ListenAndServe(":"+cfg.TrendsPort, r); err != nil {
		log.Fatal(err)
	}
}
