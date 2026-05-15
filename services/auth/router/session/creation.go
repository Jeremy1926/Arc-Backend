package router

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	api "github.com/Arc-Services/Arc/shared/api/util"
	auth "github.com/Arc-Services/Arc/shared/auth/util"
	"github.com/Arc-Services/Arc/shared/database"
	"github.com/Arc-Services/Arc/shared/database/client"
	"github.com/Arc-Services/Arc/shared/middleware"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var (
	account_db = client.New("http://account:2525")
	jwtParser  = jwt.NewParser(jwt.WithoutClaimsValidation())
)

func CreateSession(c *gin.Context) {
	shouldWeAllowMore := false

	var body struct {
		Identity string `json:"identity"` // should be a JWT, if this is anything else reject
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(500, api.ToJson(api.CreateError(c, "binding error: "+err.Error(), 500)))
		return
	}

	if body.Identity == "" {
		c.JSON(400, api.ToJson(api.CreateError(c, "identity is missing", 400)))
		return
	}

	db := middleware.MustGetClientDB(c)
	clientID := middleware.MustGetClientID(c)
	config := middleware.MustGetConfiguration(c)

	arc_auth := c.GetHeader("X-Arc-Auth")
	if arc_auth != "" && config != nil {
		api, ok := config["api"].(map[string]interface{})
		if ok {
			key, ok := api["key"].(string)
			if ok && key == arc_auth {
				shouldWeAllowMore = true
			}
		}
	} else if config != nil {
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

		if !hasBetaPermSys {
			shouldWeAllowMore = true
		}
	}

	// we should check signature here, but since i haven't setup any keys yet, skip for now

	payload, _, err := jwtParser.ParseUnverified(body.Identity, jwt.MapClaims{})
	if err != nil {
		c.JSON(400, api.ToJson(api.CreateError(c, "identity is invalid, "+err.Error(), 400)))
		return
	}

	// i should clean this up a bit tbh
	claims := payload.Claims.(jwt.MapClaims)
	iat, _ := claims.GetIssuedAt()
	exp, _ := claims.GetExpirationTime()
	iss, _ := claims.GetIssuer()
	sub, _ := claims.GetSubject()
	dn, _ := claims["dn"].(string)

	if iss == "" || sub == "" || dn == "" || iat == nil || exp == nil {
		c.JSON(400, api.ToJson(api.CreateError(c, "identity is missing required claims", 400)))
		return
	}

	account_db.ClientID = clientID

	resp, err := account_db.Get(sub)
	var account *database.Account
	if err != nil {
		if strings.Contains(err.Error(), "key not found") {
			account = &database.Account{
				DisplayName: dn,
				Country:     "US",
				Active:      true,
			}

			account_db.Set(sub, account)
			account_db.Set(fmt.Sprintf("account:%s", sub), account)
			account_db.Set(fmt.Sprintf("account:%s", strings.ToLower(dn)), account)
		} else {
			c.JSON(500, api.ToJson(api.CreateError(c, "database error: "+err.Error(), 500)))
			return
		}
	} else {
		account = &database.Account{}
		b, _ := json.Marshal(resp.Value)
		_ = json.Unmarshal(b, account)

		if account.DisplayName != dn {
			account.DisplayName = dn
			account_db.Set(sub, account)
			account_db.Set(fmt.Sprintf("account:%s", sub), account)
			account_db.Set(fmt.Sprintf("account:%s", strings.ToLower(dn)), account)
		}
	}

	// TODO: Actual permissions system - andr1ww
	var permissions []string
	if shouldWeAllowMore {
		permissions = []string{"anticheat:ws:connect", "account:read", "account:write"}
	} else {
		permissions = []string{"all:read"}
	}

	session := &database.Session{
		Owner:       sub,
		Issuer:      iss,
		Client:      clientID,
		Permissions: permissions,
	}

	token, err := auth.HashSession(session)
	if err != nil {
		c.JSON(500, api.ToJson(api.CreateError(c, "token generation error: "+err.Error(), 500)))
		return
	}

	err = db.SetTTLJSON(token, session, time.Duration(int(exp.Unix()-iat.Unix()))*time.Second)
	if err != nil {
		c.JSON(500, api.ToJson(api.CreateError(c, "failed to save session: "+err.Error(), 500)))
		return
	}

	err = db.SetTTLJSON(fmt.Sprintf("session:%s", token), session, time.Duration(int(exp.Unix()-iat.Unix()))*time.Second)
	if err != nil {
		c.JSON(500, api.ToJson(api.CreateError(c, "failed to save session: "+err.Error(), 500)))
		return
	}

	c.JSON(200, gin.H{
		"auth": gin.H{
			"token": token,
			"clid":  clientID,
		},
		"account": account,
	})
}
