package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/rainbowmga/timetravel/api"
	"github.com/rainbowmga/timetravel/config"
	"github.com/rainbowmga/timetravel/db"
	"github.com/rainbowmga/timetravel/repository"
	"github.com/rainbowmga/timetravel/service"
)

func main() {
	cfg := config.Load()

	database, err := db.New(db.Config{
		Path:            cfg.Database.Path,
		MaxOpenConns:    cfg.Database.MaxOpenConns,
		MaxIdleConns:    cfg.Database.MaxIdleConns,
		ConnMaxLifetime: cfg.Database.ConnMaxLifetime,
	})
	if err != nil {
		log.Fatalf("db init failed: %v", err)
	}
	defer database.Close()

	if err := database.Initialize(context.Background()); err != nil {
		log.Fatalf("schema init failed: %v", err)
	}
	log.Println("database ready")

	// wire up services
	repo := repository.NewSQLiteRepository(database.DB())
	v1Svc := service.NewSQLiteRecordService(repo)
	v2Svc := service.NewSQLiteVersionedService(repo)

	router := mux.NewRouter()

	// health endpoints
	router.Path("/api/v1/health").HandlerFunc(healthHandler)
	router.Path("/api/v2/health").HandlerFunc(healthHandler)

	// v1 routes (backwards compat)
	v1 := api.NewAPI(v1Svc)
	v1.CreateRoutes(router.PathPrefix("/api/v1").Subrouter())

	// v2 routes (versioned)
	v2 := api.NewAPIV2(v2Svc)
	v2.CreateRoutes(router.PathPrefix("/api/v2").Subrouter())

	srv := &http.Server{
		Handler:      router,
		Addr:         cfg.Server.Address,
		WriteTimeout: cfg.Server.WriteTimeout,
		ReadTimeout:  cfg.Server.ReadTimeout,
	}

	// start in background
	go func() {
		log.Printf("listening on %s", cfg.Server.Address)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	// wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("forced shutdown: %v", err)
	}
	log.Println("goodbye")
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}
