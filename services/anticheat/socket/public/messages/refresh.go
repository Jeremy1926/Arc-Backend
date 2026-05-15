package messages

import (
	"encoding/json"
	"time"

	anticheat "github.com/Arc-Services/Arc/services/anticheat/main"
	"github.com/Arc-Services/Arc/services/anticheat/main/classes"
	auth "github.com/Arc-Services/Arc/shared/auth/util"
	"github.com/Arc-Services/Arc/shared/database"
	"github.com/Arc-Services/Arc/shared/database/client"
	"github.com/gorilla/websocket"
)

var (
	auth_db = client.New("http://auth:8080")
)

func Refresh(client *classes.PublicClient, data map[string]interface{}) error {
	token, err := auth.UnhashSession(client.Token)
	if err != nil {
		return err
	}

	auth_db.ClientID = token.Client

	session := &database.Session{
		Owner:       token.Owner,
		Issuer:      token.Issuer,
		Client:      token.Client,
		Permissions: token.Permissions,
	}

	hashed, err := auth.HashSession(session)
	if err != nil {
		return err
	}

	iat := time.Now()
	exp := iat.Add(8 * time.Hour)

	_, err = auth_db.SetTTL(hashed, session, exp.Unix()-iat.Unix())
	if err != nil {
		return err
	}

	bytes, err := json.Marshal(map[string]interface{}{
		"type":  "refresh",
		"token": hashed,
	})

	if err != nil {
		return err
	}

	encrypted := anticheat.Encrypt(string(bytes), client.Key)
	return client.Socket.WriteMessage(websocket.BinaryMessage, encrypted)
}
