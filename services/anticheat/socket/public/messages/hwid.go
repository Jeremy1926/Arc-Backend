package messages

import (
	"encoding/json"
	"errors"
	"fmt"

	anticheat "github.com/Arc-Services/Arc/services/anticheat/main"
	"github.com/Arc-Services/Arc/services/anticheat/main/classes"
	"github.com/Arc-Services/Arc/services/anticheat/util"
	database "github.com/Arc-Services/Arc/shared/database/classes/anticheat"
	"github.com/Arc-Services/Arc/shared/database/client"
	"github.com/Arc-Services/Arc/shared/middleware"
	"github.com/gorilla/websocket"
)

var (
	account_db = client.New("http://account:2525")
)

func Hwid(client *classes.PublicClient, data map[string]interface{}) error {
	if !client.HasCompletedChallenge {
		util.RemoveClient(client)
		return errors.New("validation failed")
	}

	hwids, ok := data["hwids"].([]interface{})
	if !ok || len(hwids) == 0 {
		hwid, ok := data["hwid"].(string)
		if !ok || hwid == "" {
			util.RemoveClient(client)
			return errors.New("hwid is required")
		}

		hwids = []interface{}{hwid}
	}

	db := middleware.MustGetClientDB(client.Context)
	if db == nil {
		util.RemoveClient(client)
		return errors.New("client database is nil")
	}

	id := middleware.MustGetClientID(client.Context)

	for _, raw := range hwids {
		hwid, ok := raw.(string)
		if !ok {
			continue
		}

		if len(hwid) <= 30 { // This is temp, but its cuz like we dont want old hwids to break once this update is applied
			util.RemoveClient(client)
			return errors.New("invalid hwid length")
		}

		key := fmt.Sprintf("anticheat:private:hardware:%s:%s", hwid, client.AccountID)
		db.SetJSON(key, &database.Hardware{
			AccountID: client.AccountID,
			HWID:      hwid,
		})

		if err := db.Scan(fmt.Sprintf("anticheat:private:hardware:%s:", hwid), func(k string, v []byte) error {
			// return nil on other errors besides the main, so we dont kick the client for a server issue.
			if k == key {
				return nil
			}

			var other database.Hardware
			if err := json.Unmarshal(v, &other); err != nil {
				return nil
			}

			res, err := account_db.GetFor(id, other.AccountID)
			if err != nil {
				return nil
			}

			account, ok := res.Value.(map[string]any)
			if ok && account["active"] == true {
				return nil
			}

			return errors.New("hwid linked to inactive account: " + other.AccountID)
		}); err != nil {
			util.RemoveClient(client)
			return err
		}
	}

	client.Ready.Store(true)
	return client.Socket.WriteMessage(websocket.BinaryMessage, anticheat.Encrypt(`{"type":"ready"}`, client.Key))
}
