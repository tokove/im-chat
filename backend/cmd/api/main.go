package main

import (
	"backend/internal/config"
	"backend/internal/db"
	"backend/internal/router"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

//  go run ./cmd/api/ -config ./config/dev.env

func main() {
	// init config 
	cfg := config.LoadConfig()

	// init db
	db.InitDB(cfg.DBPath, cfg.DBName)
	defer db.Close()

	// init server
	server := &http.Server{
		Addr:    cfg.HTTPServer.Address,
		Handler: nil,
	}

	r := router.SetupRouter()
	server.Handler = r

	// shutdown gracefully
	shutdownCh := make(chan os.Signal, 1)
	signal.Notify(shutdownCh, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("Server is running: http://%s", cfg.HTTPServer.Address)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to run server, err: %v", err)
		}
	}()

	sig := <-shutdownCh
	log.Printf("Received shutdown signal: %v", sig)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown failed, err: %v", err)
	}

	log.Println("Server shutdown successfully")
}
