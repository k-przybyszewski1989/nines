package main

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/nines/backend/internal/api"
	"github.com/nines/backend/internal/config"
	"github.com/nines/backend/internal/db"
	"github.com/nines/backend/internal/ws"
	"github.com/sirupsen/logrus"
)

func main() {
	cfg, err := config.Load(".env")
	if err != nil {
		logrus.Fatalf("load config: %v", err)
	}

	database, err := db.Connect(cfg.DSN())
	if err != nil {
		logrus.Fatalf("connect to db: %v", err)
	}
	if err := db.Migrate(database); err != nil {
		logrus.Fatalf("migrate: %v", err)
	}

	wsManager := ws.NewManager()
	h := &api.Handler{DB: database, WSManager: wsManager}

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
		apiGroup.POST("/games/join", h.JoinGame)
		apiGroup.POST("/games/:id/move", h.MakeMove)
	}

	r.GET("/ws/:gameId", ws.ServeWS(wsManager, database))

	logrus.Infof("listening on :%s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		logrus.Fatalf("run: %v", err)
	}
}
