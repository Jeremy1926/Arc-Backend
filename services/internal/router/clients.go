package router

import (
	api "github.com/Arc-Services/Arc/shared/api/util"
	"github.com/Arc-Services/Arc/shared/database"
	"github.com/gin-gonic/gin"
)

func ListClients(c *gin.Context) {
	clients, err := database.ListClients()
	if err != nil {
		c.JSON(500, api.ToJson(api.CreateError(c, err.Error(), 500)))
		return
	}
	c.JSON(200, gin.H{"clients": clients})
}

func RegisterClient(c *gin.Context) {
	var req struct {
		ID   string `json:"id" binding:"required"`
		Name string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, api.ToJson(api.CreateError(c, err.Error(), 400)))
		return
	}

	if err := database.RegisterClient(req.ID, req.Name); err != nil {
		c.JSON(500, api.ToJson(api.CreateError(c, err.Error(), 500)))
		return
	}

	c.JSON(200, gin.H{"message": "client registered successfully"})
}

func DeactivateClient(c *gin.Context) {
	clientID := c.Param("id")

	if err := database.DeactivateClient(clientID); err != nil {
		c.JSON(500, api.ToJson(api.CreateError(c, err.Error(), 500)))
		return
	}

	c.JSON(200, gin.H{"message": "client deactivated successfully"})
}

func ReactivateClient(c *gin.Context) {
	clientID := c.Param("id")

	if err := database.ReactivateClient(clientID); err != nil {
		c.JSON(500, api.ToJson(api.CreateError(c, err.Error(), 500)))
		return
	}

	c.JSON(200, gin.H{"message": "client reactivated successfully"})
}
