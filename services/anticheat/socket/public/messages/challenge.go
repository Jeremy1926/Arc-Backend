package messages

import (
	"encoding/hex"
	"fmt"

	anticheat "github.com/Arc-Services/Arc/services/anticheat/main"
	"github.com/Arc-Services/Arc/services/anticheat/main/classes"
	"github.com/Arc-Services/Arc/services/anticheat/util"
	"github.com/gorilla/websocket"
)

func Challenge(client *classes.PublicClient, data map[string]interface{}) error {
	response, ok := data["res"].(string)
	isValid := client.Challenge != nil &&
		ok &&
		response == hex.EncodeToString(client.Challenge)

	if isValid {
		client.HasCompletedChallenge = true
		sendRes(client, "challengeSuccess")
		return nil
	}

	sendRes(client, "challengeFailed")
	util.RemoveClient(client)
	return nil
}

func sendRes(client *classes.PublicClient, messageType string) error {
	msg := fmt.Sprintf(`{"type":"%s"}`, messageType)
	encrypted := anticheat.Encrypt(msg, client.Key)
	return client.Socket.WriteMessage(websocket.BinaryMessage, encrypted)
}
