package private

import (
	"crypto/rand"
	"encoding/hex"
	"net"

	"github.com/Arc-Services/Arc/services/management/util"
	api "github.com/Arc-Services/Arc/shared/api/util"
	"github.com/Arc-Services/Arc/shared/database"
	"github.com/Arc-Services/Arc/shared/middleware"
	"github.com/gin-gonic/gin"
	"github.com/ipinfo/go/v2/ipinfo"
	"golang.org/x/crypto/argon2"
)

var (
	client = ipinfo.NewClient(nil, nil, "")
)

func Register(c *gin.Context) {
	var body struct {
		Email               string   `json:"email"`
		Password            string   `json:"password"`
		Name                string   `json:"name"`
		Permissions         []string `json:"permissions"`
		ForceDeleteExisting bool     `json:"force_del"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, api.ToJson(api.CreateError(c, "invalid request body", 400)))
		return
	}

	db := middleware.MustGetClientDB(c)

	b := make([]byte, 16)
	rand.Read(b)
	accId := hex.EncodeToString(b)

	info, err := client.GetIPInfo(net.ParseIP(c.ClientIP()))
	if err != nil {
		c.JSON(500, api.ToJson(api.CreateError(c, "failed to get country: "+err.Error(), 500)))
		return
	}

	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		c.JSON(500, api.ToJson(api.CreateError(c, "failed to get salt: "+err.Error(), 500)))
		return
	}

	hash := argon2.IDKey([]byte(body.Password), salt, 1, 64*1024, 4, 32)
	hashedPassword := hex.EncodeToString(salt) + ":" + hex.EncodeToString(hash)

	manager := &database.Manager{
		AccountId:   accId,
		Email:       body.Email,
		Password:    hashedPassword,
		DisplayName: body.Name,
		Permissions: body.Permissions,
		Country:     info.Country,
	}

	keys := []string{
		"mng:email:" + body.Email,
		"mng:id:" + accId,
	}

	for _, key := range keys {
		exists, err := db.Get(key)
		if err != nil {
			continue
		}
		if exists != nil {
			if !body.ForceDeleteExisting {
				c.JSON(400, api.ToJson(api.CreateError(c, "manager already exists", 400)))
				return
			} else {
				db.Delete(key)
			}
		}
	}

	for _, key := range keys {
		if err := db.SetJSON(key, manager); err != nil {
			c.JSON(500, api.ToJson(api.CreateError(c, "database error: "+err.Error(), 500)))
			return
		}
	}

	util.IndexManager(body.Email, middleware.MustGetClientID(c))

	c.JSON(200, gin.H{"msg": "okay!"})
}
