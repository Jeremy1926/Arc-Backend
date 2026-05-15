package public

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	anticheat "github.com/Arc-Services/Arc/services/anticheat/main"
	"github.com/Arc-Services/Arc/services/anticheat/main/classes"
	"github.com/Arc-Services/Arc/services/anticheat/socket/public/messages"
	"github.com/gorilla/websocket"
)

type MessageHandler func(*classes.PublicClient, map[string]interface{}) error

var messageHandlers = map[string]MessageHandler{
	"challenge":     messages.Challenge,
	"detection":     messages.Detection,
	"pong":          messages.Pong,
	"client_info":   messages.ClientInfo,
	"configuration": messages.Configuration,
	"refresh":       messages.Refresh,
	"crash":         messages.Crash,
	"hwid":          messages.Hwid,
	"travel":        messages.Travel,
}

func HandleMessage(client *classes.PublicClient, message []byte) error {
	if client.Challenge == nil {
		if len(message) != 16 {
			return fmt.Errorf("invalid message length for new client")
		}

		for i := 1; i < 16; i++ {
			message[i] ^= (0x67 & message[i-1]) ^ 0x41
		}

		hmacBuffer := make([]byte, 16)
		copy(hmacBuffer, message[:16])

		client.Key = hmacBuffer

		challenge := make([]byte, 32)
		rand.Read(challenge)

		hex := hex.EncodeToString(challenge)
		encrypted := anticheat.Encrypt(fmt.Sprintf("{\"challenge\":\"%s\"}", hex), client.Key)
		client.Socket.WriteMessage(websocket.BinaryMessage, encrypted)

		challenge[0] ^= 0xaf
		challenge[0] &= 0xfa
		challenge[0] |= 0x9
		challenge[len(challenge)-1] &= 0xf5
		challenge[len(challenge)-1] ^= ^challenge[0]

		for i := 1; i < len(challenge)-1; i++ {
			challenge[i] ^= challenge[i-1]
			challenge[i] ^= challenge[i+1]
		}

		client.Challenge = challenge
		return nil
	}

	if len(message) <= 32 {
		return fmt.Errorf("message too short")
	}

	dataLen := len(message) - 48
	dataBuffer := message[:dataLen]
	ivBuffer := message[dataLen : dataLen+16]
	sigBuffer := message[dataLen+16:]

	decrypted, err := anticheat.Decrypt(dataBuffer, ivBuffer)
	if err != nil {
		return err
	}

	signature := anticheat.Sign(decrypted, client.Key)
	if hex.EncodeToString(signature) != hex.EncodeToString(sigBuffer) {
		return fmt.Errorf("signature verification failed")
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(strings.ReplaceAll(decrypted, "\\", "\\\\")), &data); err != nil {
		return err
	}

	//currTime := time.Now().UTC()
	//msgTimeF, ok := data["time"].(float64)

	//if ok && msgTimeF > 0 && (currTime.Add(time.Duration(-24)*time.Hour).Unix() > int64(msgTimeF) || int64(msgTimeF) >= currTime.Add(time.Duration(24)*time.Hour).Unix()) {
	//	return fmt.Errorf("message is missing time, expired, or not yet valid (T: %d, MT: %d)", currTime.Unix(), int64(msgTimeF))
	//}

	msgType, ok := data["type"].(string)
	if !ok {
		return fmt.Errorf("invalid type field")
	}

	handler, exists := messageHandlers[msgType]
	if !exists {
		return fmt.Errorf("unknown message type: %s", msgType)
	}

	return handler(client, data)
}

func HandleClose(conn *websocket.Conn) {
	anticheat.RemoveClientByConn(conn)
	conn.Close()
}
