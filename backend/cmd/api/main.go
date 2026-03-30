package main

import (
	"backend/internal/config"
	"backend/internal/db"
	"backend/internal/realtime"
	"backend/internal/router"
	"backend/pkg/utils"
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
	
	// new hub
	hub := realtime.NewHub()

	// init server
	r := router.SetupRouter(hub)
	server := &http.Server{
		Addr:    cfg.HTTPServer.Address,
		Handler: r,
	}

	// init jwtKey
	utils.InitJWTKey(cfg.JWTKey)

	go func() {
		log.Printf("Server is running: http://%s", cfg.HTTPServer.Address)
		// Health Checks
		log.Printf("Health Check HTTP, GET: http://%s/api/health-check-http", cfg.HTTPServer.Address)
		log.Printf("Health Check Websocket, GET: ws://%s/api/health-check-ws", cfg.HTTPServer.Address)
		// Auths
		log.Printf("Email Register, POST: http://%s/api/auth/register-email", cfg.HTTPServer.Address)
		log.Printf("Email Login, POST: http://%s/api/auth/login-email", cfg.HTTPServer.Address)
		log.Printf("Logout, POST: http://%s/api/auth/logout", cfg.HTTPServer.Address)
		log.Printf("Refresh Session, POST: http://%s/api/auth/refresh-session", cfg.HTTPServer.Address)
		log.Printf("Get Current User, POST: http://%s/api/auth/current-user", cfg.HTTPServer.Address)
		// Users
		log.Printf("GetUserByID, POST: http://%s/api/users/:id", cfg.HTTPServer.Address)
		// Conversations
		log.Printf("GET Conversation, GET: http://%s/api/conversations/privates/:private_id", server.Addr)
		log.Printf("Create Conversation, POST: http://%s/api/conversations/privates/create", server.Addr)
		log.Printf("GET All Conversations: http://%s/api/conversations", server.Addr)
		log.Printf("GET Conversation Messages (paginated): http://%s/api/conversations/privates/:private_id/messages?page=1&limit=20", server.Addr)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to run server, err: %v", err)
		}
	}()

	// shutdown gracefully
	shutdownCh := make(chan os.Signal, 1)
	signal.Notify(shutdownCh, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	sig := <-shutdownCh
	log.Printf("Received shutdown signal: %v", sig)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown failed, err: %v", err)
	}

	hub.Shutdown()

	signal.Stop(shutdownCh) // 通知所有信号停止向chan发送信号
	close(shutdownCh)

	log.Println("Server shutdown successfully")
}
