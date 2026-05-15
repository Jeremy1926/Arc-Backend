package router

import (
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/Arc-Services/Arc/services/management/util"
	api "github.com/Arc-Services/Arc/shared/api/util"
	auth "github.com/Arc-Services/Arc/shared/auth/util"
	"github.com/Arc-Services/Arc/shared/database"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/argon2"
)

func Login(c *gin.Context) {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, api.ToJson(api.CreateError(c, "invalid request body", 400)))
		return
	}

	clid, ok := util.LookupManagerClient(body.Email)
	if !ok {
		c.JSON(401, api.ToJson(api.CreateError(c, "invalid email", 401)))
		return
	}

	db, err := database.GetClientDB(clid)
	if err != nil {
		c.JSON(500, api.ToJson(api.CreateError(c, "database error: "+err.Error(), 500)))
		return
	}

	var manager database.Manager
	err = db.GetJSON(fmt.Sprintf("mng:email:%s", body.Email), &manager)
	if err != nil {
		c.JSON(401, api.ToJson(api.CreateError(c, "invalid email", 401)))
		return
	}

	parts := strings.Split(manager.Password, ":")
	if len(parts) != 2 {
		c.JSON(500, api.ToJson(api.CreateError(c, "invalid password format", 500)))
		return
	}

	salt, err := hex.DecodeString(parts[0])
	if err != nil {
		c.JSON(500, api.ToJson(api.CreateError(c, "invalid password format", 500)))
		return
	}

	expectedHash, err := hex.DecodeString(parts[1])
	if err != nil {
		c.JSON(500, api.ToJson(api.CreateError(c, "invalid password format", 500)))
		return
	}

	attemptHash := argon2.IDKey([]byte(body.Password), salt, 1, 64*1024, 4, 32)
	if subtle.ConstantTimeCompare(attemptHash, expectedHash) != 1 {
		c.JSON(401, api.ToJson(api.CreateError(c, "invalid password", 401)))
		return
	}

	session := &database.Session{
		Owner:       manager.AccountId,
		Issuer:      "arc-management",
		Client:      clid,
		Permissions: manager.Permissions,
	}

	metadata, err := database.GetClientMetadata(clid)
	if err != nil {
		c.JSON(500, api.ToJson(api.CreateError(c, "failed to retrieve client metadata: "+err.Error(), 500)))
		return
	}

	iat := time.Now()
	exp := iat.Add(14 * 24 * time.Hour)
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

	c.JSON(200, gin.H{
		"token": token,
		"client": gin.H{
			"id":   clid,
			"name": metadata.Name,
		},
		"permissions": manager.Permissions,
	})
}
