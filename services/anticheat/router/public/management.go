package router

import (
	"encoding/json"
	"fmt"
	"strings"

	anticheat "github.com/Arc-Services/Arc/services/anticheat/main"
	"github.com/Arc-Services/Arc/services/anticheat/socket/public"
	"github.com/Arc-Services/Arc/shared/cache"
	"github.com/Arc-Services/Arc/shared/cast"
	"github.com/Arc-Services/Arc/shared/database"
	"github.com/Arc-Services/Arc/shared/database/client"
	"github.com/Arc-Services/Arc/shared/middleware"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var (
	account_db = client.New("http://account:2525")
)

func Management(c *gin.Context) {
	auth := c.GetHeader("X-Arc-Auth")
	if auth == "" {
		c.Status(401)
		return
	}

	config := middleware.MustGetConfiguration(c)
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
	}

	action := c.Query("action")

	switch action {
	case "all_hardware":
		{
			var hardwares []database.Hardware
			db := middleware.MustGetClientDB(c)
			if db == nil {
				c.Status(500)
				return
			}
			if err := db.Scan(fmt.Sprintf("anticheat:private:hardware"), func(k string, v []byte) error {
				var other database.Hardware
				if err := json.Unmarshal(v, &other); err != nil {
					return err
				}

				hardwares = append(hardwares, other)
				return nil
			}); err != nil {
				c.Status(500)
				return
			}

			c.JSON(200, gin.H{"hardwares": hardwares})
			return
		}
	case "active":
		var body struct {
			AccountId string `json:"account_id"`
			Active    bool   `json:"active"`
		}

		if err := c.BindJSON(&body); err != nil {
			c.Status(400)
			return
		}

		id := middleware.MustGetClientID(c)
		res, err := account_db.GetFor(id, body.AccountId)
		if err != nil {
			c.Status(500)
			return
		}

		m, ok := res.Value.(map[string]any)
		if ok {
			account := &database.Account{
				DisplayName: cast.Str(m["display_name"]),
				Country:     cast.Str(m["country"]),
				Active:      body.Active,
			}

			account_db.ClientID = id
			account_db.Set(body.AccountId, account)
			account_db.Set(fmt.Sprintf("account:%s", body.AccountId), account)
			account_db.Set(fmt.Sprintf("account:%s", strings.ToLower(account.DisplayName)), account)
		}

		c.JSON(200, gin.H{"status": "updated"})
		return

	case "message":
		var body struct {
			Message     map[string]interface{} `json:"message"`
			DisplayName string                 `json:"display_name"`
			AccountId   string                 `json:"account_id"`
		}

		if err := c.BindJSON(&body); err != nil {
			c.Status(400)
			return
		}

		Clients := anticheat.GetClients(middleware.MustGetClientID(c))
		for _, p := range Clients {
			if p.AccountID == body.AccountId {
				bytes, _ := json.Marshal(body.Message)
				encrypted := anticheat.Encrypt(string(bytes), p.Key)
				p.Socket.WriteMessage(websocket.BinaryMessage, encrypted)
				c.JSON(200, gin.H{"status": "sent"})
				return
			}

			if p.DisplayName == body.DisplayName {
				bytes, _ := json.Marshal(body.Message)
				encrypted := anticheat.Encrypt(string(bytes), p.Key)
				p.Socket.WriteMessage(websocket.BinaryMessage, encrypted)
				c.JSON(200, gin.H{"status": "sent"})
				return
			}
		}
	case "kick":
		var body struct {
			DisplayName string `json:"display_name"`
			AccountId   string `json:"account_id"`
		}

		if err := c.BindJSON(&body); err != nil {
			c.Status(400)
			return
		}

		Clients := anticheat.GetClients(middleware.MustGetClientID(c))
		for _, p := range Clients {
			if p.AccountID == body.AccountId {
				cache.Delete(p.CacheKey)
				public.HandleClose(p.Socket)
				c.JSON(200, gin.H{"status": "kicked"})
				return
			}

			if p.DisplayName == body.DisplayName {
				cache.Delete(p.CacheKey)
				public.HandleClose(p.Socket)
				c.JSON(200, gin.H{"status": "kicked"})
				return
			}
		}
	}

	c.JSON(200, gin.H{"status": "invalid"})
}
