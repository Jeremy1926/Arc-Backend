package messages

import (
	"encoding/json"

	anticheat "github.com/Arc-Services/Arc/services/anticheat/main"
	"github.com/Arc-Services/Arc/services/anticheat/main/classes"
	"github.com/gorilla/websocket"
)

func ClientInfo(client *classes.PublicClient, data map[string]interface{}) error {
	// bytes, ok := data["info"].(string)
	// if !ok {
	// 	return nil
	// }

	// info := make(map[string]interface{})
	// if err := json.Unmarshal([]byte(bytes), &info); err != nil {
	// 	return err
	// }

	payload, err := json.Marshal(map[string]interface{}{
		"type": "client_info",
		"info": map[string]interface{}{
			"version": client.Version, //anticheat.Version,
		},
	})
	if err != nil {
		return err
	}

	encrypted := anticheat.Encrypt(string(payload), client.Key)
	return client.Socket.WriteMessage(websocket.BinaryMessage, encrypted)
}
