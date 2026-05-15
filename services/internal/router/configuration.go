package router

import (
	api "github.com/Arc-Services/Arc/shared/api/util"
	"github.com/Arc-Services/Arc/shared/database"
	"github.com/gin-gonic/gin"
)

func ConfigureClient(c *gin.Context) {
	clientID := c.Param("id")

	var req struct {
		Configuration map[string]interface{} `json:"configuration" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if err := database.ConfigureClient(clientID, req.Configuration); err != nil {
		c.JSON(500, api.ToJson(api.CreateError(c, err.Error(), 500)))
		return
	}

	c.JSON(200, gin.H{"message": "client configured successfully"})
}
