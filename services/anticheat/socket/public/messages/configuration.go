package messages

import (
	"encoding/json"

	anticheat "github.com/Arc-Services/Arc/services/anticheat/main"
	"github.com/Arc-Services/Arc/services/anticheat/main/classes"
	"github.com/gorilla/websocket"
)

func Configuration(client *classes.PublicClient, data map[string]interface{}) error {
	cfg := make(map[string]interface{}, len(client.Configuration))
	for k, v := range client.Configuration {
		cfg[k] = v
	}

	delete(cfg, "discord")

	bytes, err := json.Marshal(map[string]interface{}{
		"type":    "configuration",
		"payload": cfg,
	})
	if err != nil {
		return err
	}

	encrypted := anticheat.Encrypt(string(bytes), client.Key)
	return client.Socket.WriteMessage(websocket.BinaryMessage, encrypted)
}
