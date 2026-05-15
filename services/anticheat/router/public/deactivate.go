package router

import (
	"fmt"
	"strings"

	anticheat "github.com/Arc-Services/Arc/services/anticheat/main"
	"github.com/Arc-Services/Arc/services/anticheat/socket/public"
	"github.com/Arc-Services/Arc/shared/cache"
	"github.com/Arc-Services/Arc/shared/cast"
	"github.com/Arc-Services/Arc/shared/database"
	"github.com/Arc-Services/Arc/shared/middleware"
	"github.com/gin-gonic/gin"
)

func Deactivate(c *gin.Context) {
	auth := c.GetHeader("X-Arc-Auth")
	if auth == "" {
		c.Status(401)
		return
	}

	id := middleware.MustGetClientID(c)
	config := middleware.MustGetConfiguration(c)
	accid := c.Param("id")
	res, err := account_db.GetFor(id, accid)
	if err != nil {
		c.Status(500)
		return
	}

	api, ok := config["api"].(map[string]interface{})
	if ok {
		key, ok := api["key"].(string)
		if ok && key != auth {
			c.Status(401)
			return
		} else if !ok {
			c.Status(401)
			return
		}
	} else {
		c.Status(401)
		return
	}

	m, ok := res.Value.(map[string]any)
	if ok {
		account := &database.Account{
			DisplayName: cast.Str(m["display_name"]),
			Country:     cast.Str(m["country"]),
			Active:      false,
		}

		account_db.ClientID = id
		account_db.Set(accid, account)
		account_db.Set(fmt.Sprintf("account:%s", accid), account)
		account_db.Set(fmt.Sprintf("account:%s", strings.ToLower(account.DisplayName)), account)
	}

	Clients := anticheat.GetClients(id)
	for _, p := range Clients {
		if p.AccountID == accid {
			cache.Delete(p.CacheKey)
			public.HandleClose(p.Socket)
			break
		}
	}

	c.Status(200)
}
