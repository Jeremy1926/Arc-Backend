package router

import (
	"github.com/Arc-Services/Arc/services/management/router/public/handlers/get"
	"github.com/Arc-Services/Arc/services/management/router/public/handlers/post"
	"github.com/gin-gonic/gin"
)

var getHandlers = map[string]func(*gin.Context){
	"accounts":      get.Accounts,
	"detections":    get.Detections,
	"configuration": get.Configuration,
}

var postHandlers = map[string]func(*gin.Context){
	"configure":       post.Configuration,
	"detections_pacc": post.DetectionPerAccount,
	"manage_acc":      post.ManageAnAccount,
}

var methods = map[string]map[string]func(*gin.Context){
	"GET":  getHandlers,
	"POST": postHandlers,
}

func Execute(c *gin.Context) {
	action := c.Query("action")
	if action == "" {
		c.Status(400)
		return
	}

	actions, ok := methods[c.Request.Method]
	if !ok {
		c.Status(405)
		return
	}

	handler, ok := actions[action]
	if !ok {
		c.Status(404)
		return
	}

	handler(c)
}
