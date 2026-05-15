package util

import (
	"fmt"
	"log"
	"runtime"

	anticheat "github.com/Arc-Services/Arc/services/anticheat/main"
	"github.com/Arc-Services/Arc/services/anticheat/main/classes"
	"github.com/Arc-Services/Arc/shared/cache"
)

func RemoveClient(target *classes.PublicClient) {
	if target.Closed.Swap(true) {
		return // already closed
	}

	pc := make([]uintptr, 10)
	n := runtime.Callers(2, pc)
	frame, _ := runtime.CallersFrames(pc[:n]).Next()
	log.Printf("[util.RemoveClient] %s called for %s (account: %s)", frame.Function, target.DisplayName, target.AccountID)

	if target.Configuration != nil {
		api, ok := target.Configuration["api"].(map[string]interface{})
		if ok {
			backend, ok := api["outgoing"].(string)
			if ok && backend != "" {
				go SendJSON(target.Configuration, fmt.Sprintf("%s/outgoing/v1/anticheat/private/revoke", backend), map[string]interface{}{
					"id": target.AccountID,
				})
			}
		}
	}

	if target.Socket != nil {
		_ = target.Socket.Close()
	}
	anticheat.RemoveClientByConn(target.Socket)
	cache.Delete(target.CacheKey)
}
