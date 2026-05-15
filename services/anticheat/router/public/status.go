package router

import (
	"net/url"
	"strings"

	anticheat "github.com/Arc-Services/Arc/services/anticheat/main"
	"github.com/Arc-Services/Arc/shared/middleware"
	"github.com/gin-gonic/gin"
)

func IsPlayerOnAntiCheat(c *gin.Context) {
	token := c.GetHeader("X-Arc-Auth")
	if token == "" {
		c.Status(401)
		return
	}

	Client, ok := anticheat.GetClientByToken(token)
	if ok && Client.Ready.Load() {
		c.Status(200)
		return
	}

	c.Status(404)
}

func PreLogin(c *gin.Context) {
	auth := c.GetHeader("X-Arc-Auth")
	api, _ := middleware.MustGetConfiguration(c)["api"].(map[string]interface{})
	if key, ok := api["key"].(string); !ok || key != auth {
		c.Status(401)
		return
	}

	data, err := c.GetRawData()
	if err != nil {
		c.Status(400)
		return
	}

	body := string(data)
	if i := strings.Index(body, "?"); i != -1 {
		body = body[i+1:]
	}
	body = strings.ReplaceAll(body, "?", "&")

	values, _ := url.ParseQuery(body)
	token := values.Get("ArcToken")
	if values.Get("Platform") != "WIN" || token == "" {
		c.Status(403)
		return
	}

	if client, ok := anticheat.GetClientByToken(token); ok &&
		client.Ready.Load() &&
		client.IsTraveling.Load() {

		client.IsTraveling.Store(false)
		c.Status(200)
		return
	}

	c.Status(404)
}

func VerifyTravel(c *gin.Context) {
	token := c.GetHeader("X-Arc-Auth")
	if token == "" {
		c.Status(401)
		return
	}

	Client, ok := anticheat.GetClientByToken(token)
	if ok && Client.Ready.Load() && Client.IsTraveling.Load() {

		/*var body struct {
			IP   string `json:"ip"`
			Port int32  `json:"port"`
		}

		if err := c.ShouldBindJSON(&body); err == nil {
			travelIP := Client.TravelIP.Load()
			travelPort := Client.TravelPort.Load()

			if travelIP == nil || body.Port == 0 || body.IP != travelIP || body.Port != travelPort {
				Client.IsTraveling.Store(false)
				c.Status(404)
				return
			}
		}*/

		Client.IsTraveling.Store(false)
		c.Status(200)
		return
	}

	Client.IsTraveling.Store(false)
	c.Status(404)
}
