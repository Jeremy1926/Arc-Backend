package api

import (
	api "github.com/Arc-Services/Arc/shared/api/models"
	"github.com/Arc-Services/Arc/shared/date"
	"github.com/gin-gonic/gin"
)

func CreateError(ctx *gin.Context, message string, code int) *api.Error {
	return &api.Error{
		Message: message,
		Code:    code,
		Route:   ctx.FullPath(),
	}
}

func ToJson(err *api.Error) gin.H {
	return gin.H{
		"message":   err.Message,
		"code":      err.Code,
		"route":     err.Route,
		"timestamp": date.RFC3339(),
	}
}
