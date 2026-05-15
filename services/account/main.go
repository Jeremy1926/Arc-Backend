package main

import (
	"encoding/hex"
	"log"
	"os"

	auth "github.com/Arc-Services/Arc/shared/auth/middleware"
	"github.com/Arc-Services/Arc/shared/cache"
	"github.com/Arc-Services/Arc/shared/database"
	"github.com/Arc-Services/Arc/shared/database/client"
	"github.com/Arc-Services/Arc/shared/middleware"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	signingKey, err := hex.DecodeString(os.Getenv("ARC_SIGNING_KEY"))
	if err != nil || len(signingKey) == 0 {
		log.Fatalf("ARC_SIGNING_KEY environment variable is missing or invalid")
	}
	client.InitSigning(signingKey)

	if err := database.InitManager("data/arc"); err != nil {
		log.Fatalf("failed to initialize database manager: %v", err)
	}
	defer database.CloseAll()

	wow := r.Group("/test")
	wow.Use(middleware.DataBase())
	wow.Use(auth.Validation())
	wow.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	client.Setup(r)
	cache.Init()

	r.Run(":2525")
}
