package main

import (
	"better-media/internal/api"
	"better-media/internal/config"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	cfg := config.Load()

	server, err := api.NewServer(cfg)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}
	defer server.Close()

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		log.Println("Shutting down server...")
		server.Close()
		os.Exit(0)
	}()

	if err := server.Run(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
