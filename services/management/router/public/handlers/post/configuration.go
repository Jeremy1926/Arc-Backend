package post

import (
	"strings"

	api "github.com/Arc-Services/Arc/shared/api/util"
	"github.com/Arc-Services/Arc/shared/database"
	"github.com/Arc-Services/Arc/shared/middleware"
	"github.com/gin-gonic/gin"
)

func Configuration(c *gin.Context) {
	var body struct {
		Paks []struct {
			Name string `json:"name"`
			Size int64  `json:"size"`
		} `json:"paks"`
		Fortnite struct {
			PatchIris        bool `json:"patch_iris"`
			EnableArenaHUD   bool `json:"enable_arena_hud"`
			PatchCorruptData bool `json:"patch_corrupt_data"`
			PatchIOStore     bool `json:"patch_iostore"`
			PatchVivox       bool `json:"patch_vivox"`
		} `json:"fortnite"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, api.ToJson(api.CreateError(c, "invalid request body: "+err.Error(), 400)))
		return
	}

	config := middleware.MustGetConfiguration(c)
	if config == nil {
		config = make(map[string]interface{})
	}

	for i, pak := range body.Paks {
		if pak.Name == "" || pak.Size <= 0 {
			c.JSON(400, api.ToJson(api.CreateError(c, "invalid pak entry at index "+string(i), 400)))
			return
		}

		pak.Name = strings.Replace(pak.Name, ".pak", "", 1)
	}

	config["whitelisted_paks"] = body.Paks
	fortniteConfig := make(map[string]interface{})

	if body.Fortnite.PatchIris {
		fortniteConfig["force_iris"] = true
	}
	if body.Fortnite.EnableArenaHUD {
		fortniteConfig["enable_arena_ui"] = true
	}
	if body.Fortnite.PatchCorruptData {
		fortniteConfig["patch_corrupt_data"] = true
	}
	if body.Fortnite.PatchIOStore {
		fortniteConfig["patch_iostore"] = true
	}
	if body.Fortnite.PatchVivox {
		fortniteConfig["patch_vivox"] = true
	}

	if len(fortniteConfig) > 0 {
		config["fortnite"] = fortniteConfig
	}

	if err := database.ConfigureClient(middleware.MustGetClientID(c), config); err != nil {
		c.JSON(500, api.ToJson(api.CreateError(c, err.Error(), 500)))
		return
	}

	c.JSON(200, gin.H{"message": "configuration updated successfully"})
}
