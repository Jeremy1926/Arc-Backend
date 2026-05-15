package messages

import (
	"github.com/Arc-Services/Arc/services/anticheat/main/classes"
)

func Travel(client *classes.PublicClient, data map[string]interface{}) error {
	client.IsTraveling.Store(true)

	ip, ok := data["ip"].(string)
	if ok {
		client.TravelIP.Store(ip)
	}

	port, ok := data["port"].(float64)
	if ok {
		client.TravelPort.Store(int32(port))
	}

	return nil
}
