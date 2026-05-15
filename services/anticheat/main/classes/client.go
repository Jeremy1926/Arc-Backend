package classes

import (
	"fmt"
	"sync/atomic"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type PublicClient struct {
	Socket                *websocket.Conn
	AccountID             string
	Token                 string
	DisplayName           string
	Key                   []byte
	Challenge             []byte
	HasCompletedChallenge bool
	Configuration         map[string]interface{}
	Context               *gin.Context
	LastPong              atomic.Int64
	CacheKey              string
	Version               float64
	Closed                atomic.Bool
	Ready                 atomic.Bool
	IsTraveling           atomic.Bool
	TravelIP              atomic.Value
	TravelPort            atomic.Int32
}

func NewPublicClient(session map[string]any, account map[string]any, t string, config map[string]interface{}, version float64) *PublicClient {
	return &PublicClient{
		Socket:                nil,
		AccountID:             session["owner"].(string),
		Token:                 t,
		DisplayName:           account["display_name"].(string),
		Key:                   nil,
		Challenge:             nil,
		HasCompletedChallenge: false,
		Configuration:         config,
		Context:               nil,
		CacheKey:              fmt.Sprintf("anticheat:public:%s", t),
		Version:               version,
		Closed:                atomic.Bool{},
		Ready:                 atomic.Bool{},
		IsTraveling:           atomic.Bool{},
	}
}
