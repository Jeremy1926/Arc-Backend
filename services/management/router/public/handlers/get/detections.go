package get

import (
	"encoding/json"
	"strings"

	api "github.com/Arc-Services/Arc/shared/api/util"
	"github.com/Arc-Services/Arc/shared/database/client"
	"github.com/Arc-Services/Arc/shared/middleware"
	"github.com/gin-gonic/gin"
)

var (
	anticheat_db = client.New("http://anticheat:2151")
)

func Detections(c *gin.Context) {
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

	accountMap := make(map[string]map[string]interface{}, len(accounts))

	for _, acc := range accounts {
		key, ok := acc["key"].(string)
		if !ok {
			continue
		}

		if !strings.HasPrefix(key, "account:") {
			continue
		}

		accountId := strings.TrimPrefix(key, "account:")
		accountMap[accountId] = acc
	}

	detections := []map[string]interface{}{}
	anticheat_db.ClientID = middleware.MustGetClientID(c)

	resp, err = anticheat_db.Scan("detection:", 0) // 0 = no limit
	if err != nil {
		c.JSON(500, api.ToJson(api.CreateError(c, "database error: "+err.Error(), 500)))
		return
	}

	if resp.Results == nil {
		c.JSON(500, api.ToJson(api.CreateError(c, "database error: nil results", 500)))
		return
	}

	results, ok = resp.Results.(map[string]interface{})
	if !ok {
		c.JSON(500, api.ToJson(api.CreateError(c, "invalid scan response", 500)))
		return
	}

	for key, raw := range results {
		m, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}

		// we get rid of these cuz we dont rlly want the public having this.
		delete(m, "type")
		delete(m, "info")

		accountId, ok := m["account_id"].(string)
		if !ok {
			continue
		}

		if acc, exists := accountMap[accountId]; exists {
			m["account"] = acc
		}

		m["key"] = key // dk if key will be needed but ill put it anyway
		detections = append(detections, m)
	}

	c.JSON(200, detections)
}
