package post

import (
	api "github.com/Arc-Services/Arc/shared/api/util"
	"github.com/Arc-Services/Arc/shared/database/client"
	"github.com/Arc-Services/Arc/shared/middleware"
	"github.com/gin-gonic/gin"
)

var (
	anticheat_db = client.New("http://anticheat:2151")
)

func DetectionPerAccount(c *gin.Context) {
	var body struct {
		AccountId string `json:"account_id"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, api.ToJson(api.CreateError(c, "invalid request body: "+err.Error(), 400)))
		return
	}

	detections := []map[string]interface{}{}
	anticheat_db.ClientID = middleware.MustGetClientID(c)

	resp, err := anticheat_db.Scan("detection:"+body.AccountId, 0) // 0 = no limit
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
		m, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}

		m["key"] = key // dk if key will be needed but ill put it anyway
		detections = append(detections, m)
	}

	c.JSON(200, detections)
}
