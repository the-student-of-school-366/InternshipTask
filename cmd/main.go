package main

import (
	"InternshipTask/internal/app"
	"log"
	"os"
	"time"
)

func main() {
	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" {
		log.Fatal("DATABASE_DSN env is required")
	}

	engine, err := app.New(app.Config{
		PostgresDSN:    dsn,
		ConnectTimeout: 5 * time.Second,
	})
	if err != nil {
		log.Fatalf("init app: %v", err)
	}

	if err := engine.Run(":8080"); err != nil {
		log.Fatalf("run server: %v", err)
	}
}


