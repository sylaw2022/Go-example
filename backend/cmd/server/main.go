package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sylaw/fullstack-app/internal/api"
	"github.com/sylaw/fullstack-app/internal/config"
	"github.com/sylaw/fullstack-app/internal/repository"
	"github.com/sylaw/fullstack-app/internal/service"
)

func main() {
	// 1. Load config
	cfg := config.LoadConfig()

	// 2. Initialize Dependencies
	userRepo := repository.NewInMemoryUserRepository()
	userService := service.NewUserService(userRepo)
	r := api.SetupRouter(userService)

	// 3. Setup HTTP Server
	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	// 4. Start Server Context
	serverCtx, serverStopCtx := context.WithCancel(context.Background())

	// 5. Listen for syscall signals for gracefully shutting down
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-sig
		log.Println("Shutting down server...")

		// Shutdown signal with grace period of 30 seconds
		shutdownCtx, _ := context.WithTimeout(serverCtx, 30*time.Second)

		go func() {
			<-shutdownCtx.Done()
			if shutdownCtx.Err() == context.DeadlineExceeded {
				log.Fatal("Graceful shutdown timed out.. forcing exit.")
			}
		}()

		err := server.Shutdown(shutdownCtx)
		if err != nil {
			log.Fatal(err)
		}
		serverStopCtx()
	}()

	// 6. Start server
	log.Printf("Starting server on port %s", cfg.Port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server startup failed: %v", err)
	}

	// Wait for server context to be stopped
	<-serverCtx.Done()
	log.Println("Server exitted gracefully")
}
