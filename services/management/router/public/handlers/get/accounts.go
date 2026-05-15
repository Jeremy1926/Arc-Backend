package get

import (
	"encoding/json"

	api "github.com/Arc-Services/Arc/shared/api/util"
	"github.com/Arc-Services/Arc/shared/database/client"
	"github.com/Arc-Services/Arc/shared/middleware"
	"github.com/gin-gonic/gin"
)

var (
	account_db = client.New("http://account:2525")
)

func Accounts(c *gin.Context) {
	accounts := []map[string]interface{}{}
	account_db.ClientID = middleware.MustGetClientID(c)

	resp, err := account_db.Scan("account:", 0) // 0 = no limit
	if err != nil {
		c.JSON(500, api.ToJson(api.CreateError(c, "database error: "+err.Error(), 500)))
		return
	}

	if resp.Results == nil {
		c.JSON(500, api.ToJson(api.CreateError(c, "database error: nil results", 500)))
		return
	}

	results, ok := resp.Results.(map[string]interface{})
	if !ok {
		c.JSON(500, api.ToJson(api.CreateError(c, "invalid scan response", 500)))
		return
	}

	for key, raw := range results {
		var m map[string]interface{}

		switch v := raw.(type) {
		case string:
			if err := json.Unmarshal([]byte(v), &m); err != nil {
				continue
			}
			if m == nil {
				m = make(map[string]interface{})
			}
		case map[string]interface{}:
			m = make(map[string]interface{}, len(v)+1)
			for k, val := range v {
				m[k] = val
			}
		default:
			continue
		}

		m["key"] = key
		accounts = append(accounts, m)
	}

	c.JSON(200, gin.H{"accounts": accounts})
}
