package messages

import (
	"time"

	"github.com/Arc-Services/Arc/services/anticheat/main/classes"
)

func Pong(client *classes.PublicClient, data map[string]interface{}) error {
	client.LastPong.Store(time.Now().UnixNano())
	return nil
}
