package routines

import (
	"encoding/binary"
	"fmt"
	"math/rand/v2"
	"sync"
	"time"

	anticheat "github.com/Arc-Services/Arc/services/anticheat/main"
	"github.com/Arc-Services/Arc/services/anticheat/util"
	"github.com/gorilla/websocket"
)

var (
	mu         sync.Mutex
	heartbeats = make(map[string]chan struct{})
)

func Heartbeat(clientID string) { // not sure how stable this will be
	clients := anticheat.GetClients(clientID)

	if len(clients) == 0 { // no clients = waste of goroutines
		StopHeartbeat(clientID)
		return
	}

	at := time.Now()

	for _, client := range clients {
		if !client.HasCompletedChallenge || client.Key == nil || client.Socket == nil {
			continue
		}

		var nonce [8]byte
		binary.LittleEndian.PutUint64(nonce[:], rand.Uint64())

		encrypted := anticheat.Encrypt(fmt.Sprintf(`{"ping":"%x"}`, nonce), client.Key)
		client.Socket.WriteMessage(websocket.BinaryMessage, encrypted)
	}

	go func(ID string, pinged time.Time) {
		time.Sleep(15 * time.Second)

		// we need to fresh clients cuz some might have died
		for _, client := range anticheat.GetClients(ID) {
			if !client.HasCompletedChallenge || client.Key == nil || client.Socket == nil {
				continue
			}

			lastPong := client.LastPong.Load()
			if lastPong == 0 || time.Unix(0, lastPong).Before(pinged) {
				util.RemoveClient(client)
			}
		}
	}(clientID, at)
}

func StartHeartbeat(clientID string) {
	mu.Lock()
	defer mu.Unlock()

	if _, exists := heartbeats[clientID]; exists { // phew
		return
	}

	stop := make(chan struct{})
	heartbeats[clientID] = stop

	go HeartbeatTicker(clientID, stop)
}

func HeartbeatTicker(clientID string, stop chan struct{}) {
	ticker := time.NewTicker(35 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			Heartbeat(clientID)
		case <-stop:
			return
		}
	}
}

func StopHeartbeat(clientID string) {
	mu.Lock()
	defer mu.Unlock()

	if stop, exists := heartbeats[clientID]; exists {
		close(stop)
		delete(heartbeats, clientID)
	}
}
