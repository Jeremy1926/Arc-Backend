package post

import (
	"fmt"
	"strings"

	api "github.com/Arc-Services/Arc/shared/api/util"
	"github.com/Arc-Services/Arc/shared/database/client"
	"github.com/Arc-Services/Arc/shared/middleware"
	"github.com/gin-gonic/gin"
)

var (
	account_db = client.New("http://account:2525")
)

func ManageAnAccount(c *gin.Context) {
	var body struct {
		AccountId string `json:"account_id"`
		Active    bool   `json:"active"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, api.ToJson(api.CreateError(c, "invalid request body: "+err.Error(), 400)))
		return
	}

	clientID := middleware.MustGetClientID(c)

	accountRes, err := account_db.GetFor(clientID, body.AccountId) // session.Owner = account ID
	if err != nil {
		c.JSON(500, api.ToJson(api.CreateError(c, "failed to get account, "+err.Error(), 500)))
		return
	}

	if accountRes == nil {
		c.JSON(404, api.ToJson(api.CreateError(c, "account not found", 404)))
		return
	}

	account, ok := accountRes.Value.(map[string]any)
	if !ok {
		c.JSON(500, api.ToJson(api.CreateError(c, "invalid account data: type assertion failed", 500)))
		return
	}

	account["active"] = body.Active
	account_db.ClientID = middleware.MustGetClientID(c)

	account_db.Set(body.AccountId, account)
	account_db.Set(fmt.Sprintf("account:%s", body.AccountId), account)
	account_db.Set(fmt.Sprintf("account:%s", strings.ToLower(account["display_name"].(string))), account)

	c.JSON(200, gin.H{"message": "account updated successfully"})
}
