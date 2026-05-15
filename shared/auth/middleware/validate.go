package auth

import (
	api "github.com/Arc-Services/Arc/shared/api/util"
	auth "github.com/Arc-Services/Arc/shared/auth/util"
	"github.com/Arc-Services/Arc/shared/cast"
	"github.com/Arc-Services/Arc/shared/database"
	"github.com/Arc-Services/Arc/shared/database/client"
	"github.com/Arc-Services/Arc/shared/middleware"
	"github.com/gin-gonic/gin"
)

var (
	account_db = client.New("http://account:2525")
)

func Validation() gin.HandlerFunc {
	return func(c *gin.Context) {
		Authorization := c.GetHeader("Authorization")
		if Authorization == "" {
			c.JSON(400, api.ToJson(api.CreateError(c, "authorization header is missing", 400)))
			c.Abort()
			return
		}

		session, err := auth.UnhashSession(Authorization)
		if err != nil {
			c.JSON(401, api.ToJson(api.CreateError(c, "invalid session: "+err.Error(), 401)))
			c.Abort()
			return
		}

		clientID := middleware.MustGetClientID(c)
		account_db.ClientID = clientID

		resp, err := account_db.Get(session.Owner)
		if err != nil {
			c.JSON(500, api.ToJson(api.CreateError(c, "failed to fetch account: "+err.Error(), 500)))
			c.Abort()
			return
		}

		m := resp.Value.(map[string]interface{})

		account := &database.Account{
			DisplayName: cast.Str(m["display_name"]),
			Country:     cast.Str(m["country"]),
			Active:      cast.Bool(m["active"]),
		}

		if !account.Active {
			c.JSON(403, api.ToJson(api.CreateError(c, "account is inactive", 403)))
			c.Abort()
			return
		}

		if account.DisplayName == "" {
			c.JSON(500, api.ToJson(api.CreateError(c, "account is invalid", 500)))
			c.Abort()
			return
		}

		c.Set("account", account)
		c.Set("session", session)

		c.Next()
	}
}
