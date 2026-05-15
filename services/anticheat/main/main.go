package anticheat

import (
	"sync"

	"github.com/Arc-Services/Arc/services/anticheat/main/classes"
	"github.com/gorilla/websocket"
)

var (
	publicClients = make(map[string][]*classes.PublicClient)
	clientsMu     sync.RWMutex
	Version       = 0.9
	clientByToken = make(map[string]*classes.PublicClient)
)

type Public struct {
	AccountID   string `json:"accountId"`
	DisplayName string `json:"displayName"`
}

func AddClient(clientID string, c *classes.PublicClient) {
	clientsMu.Lock()
	defer clientsMu.Unlock()

	clientByToken[c.Token] = c

	for i, existing := range publicClients[clientID] {
		if existing.DisplayName == c.DisplayName {
			publicClients[clientID][i] = c
			return
		}
	}

	publicClients[clientID] = append(publicClients[clientID], c)
}

func GetClients(clientID string) []*classes.PublicClient {
	clientsMu.RLock()
	defer clientsMu.RUnlock()

	return publicClients[clientID]
}

func GetPubClients(clientID string) []*Public {
	clientsMu.RLock()
	defer clientsMu.RUnlock()

	list := publicClients[clientID]
	out := make([]*Public, 0, len(list))

	for _, c := range list {
		out = append(out, &Public{
			AccountID:   c.AccountID,
			DisplayName: c.DisplayName,
		})
	}

	return out
}

func GetAllClients() map[string][]*Public {
	clientsMu.RLock()
	defer clientsMu.RUnlock()

	out := make(map[string][]*Public, len(publicClients))

	for clientID, list := range publicClients {
		arr := make([]*Public, 0, len(list))

		for _, c := range list {
			arr = append(arr, &Public{
				AccountID:   c.AccountID,
				DisplayName: c.DisplayName,
			})
		}

		out[clientID] = arr
	}

	return out
}

func GetClientByToken(token string) (*classes.PublicClient, bool) {
	clientsMu.RLock()
	defer clientsMu.RUnlock()

	c, ok := clientByToken[token]
	return c, ok
}

func RemoveClient(clientID string, target *classes.PublicClient) {
	clientsMu.Lock()
	defer clientsMu.Unlock()

	list, ok := publicClients[clientID]
	if !ok {
		return
	}

	for i, c := range list {
		if c == target {
			publicClients[clientID] = append(list[:i], list[i+1:]...)
			delete(clientByToken, c.Token)
			break
		}
	}

	if len(publicClients[clientID]) == 0 {
		delete(publicClients, clientID)
	}
}

func RemoveClientByConn(targetConn *websocket.Conn) {
	clientsMu.Lock()
	defer clientsMu.Unlock()

	for clientID, list := range publicClients {
		for i, c := range list {
			if c.Socket == targetConn {
				delete(clientByToken, c.Token)
				publicClients[clientID] = append(list[:i], list[i+1:]...)
				break
			}
		}

		if len(publicClients[clientID]) == 0 {
			delete(publicClients, clientID)
		}
	}
}
