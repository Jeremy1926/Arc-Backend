package main

import (
	"encoding/hex"
	"log"
	"os"

	router "github.com/Arc-Services/Arc/services/auth/router/session"
	"github.com/Arc-Services/Arc/shared/cache"
	"github.com/Arc-Services/Arc/shared/database"
	"github.com/Arc-Services/Arc/shared/database/client"
	"github.com/Arc-Services/Arc/shared/middleware"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.New()
	r.Use(cors.Default())
	r.Use(gin.Logger())

	signingKey, err := hex.DecodeString(os.Getenv("ARC_SIGNING_KEY"))
	if err != nil || len(signingKey) == 0 {
		log.Fatalf("ARC_SIGNING_KEY environment variable is missing or invalid")
	}
	client.InitSigning(signingKey)

	if err := database.InitManager("data/arc"); err != nil {
		log.Fatalf("failed to initialize database manager: %v", err)
	}
	defer database.CloseAll()

	session := r.Group("/api/v1/auth/session")
	session.Use(middleware.DataBase())
	session.POST("/", router.CreateSession)

	client.Setup(r)
	cache.Init()

	r.Run(":8080")
}
