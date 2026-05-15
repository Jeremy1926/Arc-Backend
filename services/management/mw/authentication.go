package mw

import (
	"encoding/json"

	api "github.com/Arc-Services/Arc/shared/api/util"
	"github.com/Arc-Services/Arc/shared/middleware"
	"github.com/gin-gonic/gin"
)

func Authentication() gin.HandlerFunc {
	return func(c *gin.Context) {
		Authorization := c.GetHeader("Authorization")
		if Authorization == "" {
			c.JSON(400, api.ToJson(api.CreateError(c, "authorization header is missing", 400)))
			c.Abort()
			return
		}

		db := middleware.MustGetClientDB(c)
		tokenData, err := db.Get(Authorization)
		if err != nil {
			c.JSON(401, api.ToJson(api.CreateError(c, "invalid authorization token", 401)))
			c.Abort()
			return
		}

		var token map[string]interface{}
		err = json.Unmarshal(tokenData, &token)
		if err != nil {
			c.JSON(401, api.ToJson(api.CreateError(c, "invalid authorization token", 401)))
			c.Abort()
			return
		}

		c.Set("permissions", token["permissions"])
		c.Next()
	}
}
