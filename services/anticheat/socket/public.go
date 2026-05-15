package socket

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	anticheat "github.com/Arc-Services/Arc/services/anticheat/main"
	"github.com/Arc-Services/Arc/services/anticheat/main/classes"
	"github.com/Arc-Services/Arc/services/anticheat/socket/public"
	"github.com/Arc-Services/Arc/services/anticheat/socket/public/routines"
	"github.com/Arc-Services/Arc/services/anticheat/util"
	api "github.com/Arc-Services/Arc/shared/api/util"
	auth "github.com/Arc-Services/Arc/shared/auth/util"
	"github.com/Arc-Services/Arc/shared/cache"
	"github.com/Arc-Services/Arc/shared/database/client"
	"github.com/Arc-Services/Arc/shared/middleware"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var (
	account_db = client.New("http://account:2525")
	auth_db    = client.New("http://auth:8080")
)

func PublicAnticheatSocket(c *gin.Context) {
	token := strings.TrimSpace(c.Query("t"))
	token = strings.TrimPrefix(token, `"`)
	token = strings.TrimSuffix(token, `"`)
	token = strings.Trim(token, "\"“”")

	if token == "" {
		c.JSON(400, api.ToJson(api.CreateError(c, "missing token", 400)))
		return
	}

	clientId := middleware.MustGetClientID(c)

	var client *classes.PublicClient
	key := fmt.Sprintf("anticheat:public:%s", token)

	if err := cache.Get(key, &client); err != nil {
		var session map[string]any
		sessionRes, err := auth_db.GetFor(clientId, token)
		if err != nil {
			sessionPtr, err := auth.UnhashSession(token)
			if err != nil {
				c.JSON(401, api.ToJson(api.CreateError(c, "invalid session: "+err.Error(), 401)))
				return
			}

			perms := make([]interface{}, len(sessionPtr.Permissions))
			for i, p := range sessionPtr.Permissions {
				perms[i] = p
			}

			session = map[string]any{
				"owner":       sessionPtr.Owner,
				"permissions": perms,
			}
		} else {
			if sessionRes == nil {
				c.JSON(500, api.ToJson(api.CreateError(c, "session response is nil", 500)))
				return
			}

			var ok bool
			session, ok = sessionRes.Value.(map[string]any)
			if !ok {
				c.JSON(500, api.ToJson(api.CreateError(c, "invalid session data: type assertion failed", 500)))
				return
			}
		}

		Clients := anticheat.GetClients(middleware.MustGetClientID(c))
		for _, p := range Clients {
			if p.AccountID == session["owner"].(string) {
				c.JSON(403, api.ToJson(api.CreateError(c, "session already exists", 403)))
				return
			}
		}

		if session["permissions"] == nil {
			session["permissions"] = []interface{}{}
		}

		var perms []string

		switch v := session["permissions"].(type) {
		case []interface{}:
			for _, p := range v {
				if s, ok := p.(string); ok {
					perms = append(perms, s)
				}
			}

		case []string:
			perms = v

		default:
			session["permissions"] = []interface{}{}
		}

		config := middleware.MustGetConfiguration(c)
		hasBetaPermSys := false
		beta, ok := config["beta"].([]string)
		if ok {
			for _, v := range beta {
				if v == "arc:beta:perm_sys" {
					hasBetaPermSys = true
					break
				}
			}
		}

		canWeAllowConn := false
		for _, perm := range perms {
			if perm == "anticheat:ws:connect" {
				canWeAllowConn = true
				break
			}
		}

		if !canWeAllowConn && !hasBetaPermSys { // if they don't have the permission, but also don't have the new perm system, allow them anyway but with a warning
			canWeAllowConn = true
		}

		if !canWeAllowConn {
			c.JSON(403, api.ToJson(api.CreateError(c, "permission denied: missing anticheat:ws:connect", 403)))
			return
		}

		accountRes, err := account_db.GetFor(clientId, session["owner"].(string)) // session.Owner = account ID
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

		if account["active"] != true {
			c.JSON(403, api.ToJson(api.CreateError(c, "account is not active", 403)))
			return
		}

		client = classes.NewPublicClient(session, account, token, config, anticheat.Version)
		cache.Set(key, client, 16*time.Hour) // we want a huge TTL here since the websocket session is long-lived
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(500, api.ToJson(api.CreateError(c, "failed to upgrade to websocket", 500)))
		return
	}

	client.Socket = conn
	client.Context = c
	client.LastPong.Store(time.Now().UnixNano()) // refresh cuz of reconnection
	anticheat.AddClient(clientId, client)

	if client.Version != anticheat.Version {
		cache.Set(key, client, time.Hour) // we want them to update very soon if possible.
	} else {
		client.Version = anticheat.Version
	}

	routines.StartHeartbeat(clientId) // dw it has a check to make sure if it's already running or not

	defer func() {
		util.RemoveClient(client)
	}()

	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			break
		}

		if messageType == websocket.BinaryMessage {
			if err := public.HandleMessage(client, message); err != nil {
				log.Printf("err handling message: %v", err)
				break
			}
		}

		if messageType == websocket.CloseMessage {
			break
		}
	}
}
