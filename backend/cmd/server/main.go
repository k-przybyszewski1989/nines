package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/nines/backend/internal/api"
	"github.com/nines/backend/internal/db"
)

func main() {
	dsn := dsn()
	database, err := db.Connect(dsn)
	if err != nil {
		log.Fatalf("connect to db: %v", err)
	}
	if err := db.Migrate(database); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	h := &api.Handler{DB: database}
	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},
		AllowCredentials: false,
	}))

	apiGroup := r.Group("/api")
	{
		apiGroup.POST("/games", h.CreateGame)
		apiGroup.GET("/games/:id", h.GetGame)
		apiGroup.POST("/games/:id/move", h.MakeMove)
	}

	port := env("PORT", "8080")
	log.Printf("listening on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("run: %v", err)
	}
}

func dsn() string {
	host := env("DB_HOST", "localhost")
	port := env("DB_PORT", "3306")
	user := env("DB_USER", "nines")
	pass := env("DB_PASS", "nines")
	name := env("DB_NAME", "nines")
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", user, pass, host, port, name)
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
