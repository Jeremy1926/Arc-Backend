package main

import (
	"encoding/hex"
	"log"
	"os"

	"github.com/Arc-Services/Arc/services/management/mw"
	router_priv "github.com/Arc-Services/Arc/services/management/router/private"
	router_pub "github.com/Arc-Services/Arc/services/management/router/public"
	"github.com/Arc-Services/Arc/services/management/util"
	"github.com/Arc-Services/Arc/shared/cache"
	"github.com/Arc-Services/Arc/shared/database"
	"github.com/Arc-Services/Arc/shared/database/client"
	"github.com/Arc-Services/Arc/shared/middleware"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(cors.Default())

	signingKey, err := hex.DecodeString(os.Getenv("ARC_SIGNING_KEY"))
	if err != nil || len(signingKey) == 0 {
		log.Fatalf("ARC_SIGNING_KEY environment variable is missing or invalid")
	}
	client.InitSigning(signingKey)

	if err := util.InitManagersIndex("data/arc"); err != nil {
		log.Fatalf("failed to initialize index manager: %v", err)
	}

	if err := database.InitManager("data/arc"); err != nil {
		log.Fatalf("failed to initialize database manager: %v", err)
	}
	defer database.CloseAll()

	router_public := r.Group("/router/v1/management/public")
	router_public.POST("/login", router_pub.Login)
	router_public.Use(middleware.DataBase())
	router_public.Use(mw.Authentication())
	router_public.Any("/execute", router_pub.Execute)

	router_private := r.Group("/router/v1/management/private")
	router_private.Use(middleware.DataBase())
	router_private.POST("/register", router_priv.Register)

	client.Setup(r)
	cache.Init()

	r.Run(":6842")
}
