package main

import (
	"encoding/hex"
	"fmt"
	"log"
	"os"

	anticheat "github.com/Arc-Services/Arc/services/anticheat/main"
	router "github.com/Arc-Services/Arc/services/anticheat/router/public"
	"github.com/Arc-Services/Arc/services/anticheat/socket"
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

	r.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong")
	})

	r.GET("/internal/v1/anticheat/clients", func(ctx *gin.Context) {
		ctx.JSON(200, anticheat.GetAllClients())
	})

	r.GET("/router/v1/anticheat/public/client/:client", func(ctx *gin.Context) {
		client := ctx.Param("client")
		ctx.String(200, fmt.Sprintf("%d", len(anticheat.GetPubClients(client))))
	})

	anticheat := r.Group("/socket/v1/anticheat")
	anticheat.Use(middleware.DataBase())
	anticheat.GET("/", socket.PublicAnticheatSocket)

	router_public := r.Group("/router/v1/anticheat/public")
	router_public.Use(middleware.DataBase())
	router_public.GET("/status", router.IsPlayerOnAntiCheat)
	router_public.POST("/pre-login", router.PreLogin)
	router_public.POST("/verifyTravel", router.VerifyTravel)
	router_public.POST("/management", router.Management)
	router_public.POST("/deactivate/:id", router.Deactivate)

	client.Setup(r)
	cache.Init()

	r.Run(":2151")
}
