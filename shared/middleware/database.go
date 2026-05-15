package middleware

import (
	api "github.com/Arc-Services/Arc/shared/api/util"
	"github.com/Arc-Services/Arc/shared/database"
	"github.com/gin-gonic/gin"
)

func DataBase() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientID := c.GetHeader("X-Arc-Client")

		if clientID == "" {
			c.JSON(400, api.ToJson(api.CreateError(c, "missing client ID header", 400)))
			c.Abort()
			return
		}

		db, err := database.GetClientDB(clientID)
		if err != nil {
			c.JSON(404, api.ToJson(api.CreateError(c, "client not found: "+err.Error(), 404)))
			c.Abort()
			return
		}

		c.Set("db", db)
		c.Set("clid", clientID)
		config, err := database.GetClientConfiguration(clientID)
		if err == nil {
			c.Set("configuration", config)
		}

		c.Next()
	}
}

// this wont be used prob, but its good to have for debug
func DataBaseWithAutoRegister(defaultName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientID := c.GetHeader("X-Arc-Client")

		if clientID == "" {
			c.JSON(400, api.ToJson(api.CreateError(c, "missing client ID header", 400)))
			c.Abort()
			return
		}

		db, err := database.GetClientDB(clientID)
		if err != nil {
			name := defaultName
			if name == "" {
				name = clientID
			}

			if regErr := database.RegisterClient(clientID, name); regErr != nil {
				c.JSON(500, api.ToJson(api.CreateError(c, "failed to register client: "+regErr.Error(), 500)))
				c.Abort()
				return
			}

			db, err = database.GetClientDB(clientID)
			if err != nil {
				c.JSON(500, api.ToJson(api.CreateError(c, "failed to get client database: "+err.Error(), 500)))
				c.Abort()
				return
			}
		}

		c.Set("db", db)
		c.Set("clid", clientID)

		c.Next()
	}
}

func GetClientDB(c *gin.Context) (*database.DB, bool) {
	db, exists := c.Get("db")
	if !exists {
		return nil, false
	}

	clientDB, ok := db.(*database.DB)
	if !ok {
		return nil, false
	}

	return clientDB, true
}

func GetClientID(c *gin.Context) (string, bool) {
	clientID, exists := c.Get("clid")
	if !exists {
		return "", false
	}

	id, ok := clientID.(string)
	if !ok {
		return "", false
	}

	return id, true
}

func MustGetClientDB(c *gin.Context) *database.DB {
	db, ok := GetClientDB(c)
	if !ok {
		panic("client database not found in context")
	}
	return db
}

func MustGetClientID(c *gin.Context) string {
	id, ok := GetClientID(c)
	if !ok {
		panic("client ID not found in context")
	}
	return id
}

func MustGetConfiguration(c *gin.Context) map[string]interface{} {
	config, exists := c.Get("configuration")
	if !exists {
		return nil
	}
	configuration, ok := config.(map[string]interface{})
	if !ok {
		return nil
	}
	return configuration
}
