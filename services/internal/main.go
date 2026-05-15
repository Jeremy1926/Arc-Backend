package main

import (
	"encoding/hex"
	"log"
	"os"

	"github.com/Arc-Services/Arc/services/internal/router"
	"github.com/Arc-Services/Arc/shared/cache"
	"github.com/Arc-Services/Arc/shared/database"
	"github.com/Arc-Services/Arc/shared/database/client"
	"github.com/gin-gonic/gin"
)

func adminKeyAuth() gin.HandlerFunc {
	adminKey := os.Getenv("ARC_ADMIN_KEY")
	if adminKey == "" {
		log.Fatalf("ARC_ADMIN_KEY environment variable is missing")
	}

	return func(c *gin.Context) {
		key := c.GetHeader("X-Arc-Admin-Key")
		if key == "" || key != adminKey {
			c.JSON(401, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}
		c.Next()
	}
}

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

	admin := r.Group("/admin/v1/clients")
	admin.Use(adminKeyAuth())
	admin.GET("/", router.ListClients)
	admin.POST("/register", router.RegisterClient)
	admin.POST("/:id/deactivate", router.DeactivateClient)
	admin.POST("/:id/reactivate", router.ReactivateClient)
	admin.POST("/:id/configure", router.ConfigureClient)
	admin.GET("/:id/configuration", func(ctx *gin.Context) {
		clientID := ctx.Param("id")
		config, err := database.GetClientConfiguration(clientID)
		if err != nil {
			ctx.JSON(500, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(200, gin.H{"configuration": config})
	})

	cache.Init()

	r.Run(":6070")
}
