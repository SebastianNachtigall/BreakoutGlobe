package main

import (
	"log"
	"os"

	"breakoutglobe/internal/config"
	"breakoutglobe/internal/server"
)

func main() {
	cfg := config.Load()
	
	srv := server.New(cfg)
	
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	
	log.Printf("Starting server on port %s", port)
	if err := srv.Start(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}