package get

import (
	"github.com/Arc-Services/Arc/shared/middleware"
	"github.com/gin-gonic/gin"
)

func Configuration(c *gin.Context) {
	config := middleware.MustGetConfiguration(c)
	if config == nil {
		c.Status(500)
		return
	}

	type Pak struct {
		Name string `json:"name"`
		Size int64  `json:"size"`
	}

	type Response struct {
		Paks     []Pak `json:"paks"`
		Fortnite struct {
			PatchIris        bool `json:"patch_iris"`
			EnableArenaHUD   bool `json:"enable_arena_hud"`
			PatchCorruptData bool `json:"patch_corrupt_data"`
			PatchIOStore     bool `json:"patch_iostore"`
			PatchVivox       bool `json:"patch_vivox"`
		} `json:"fortnite"`
	}

	var resp Response

	if raw, ok := config["whitelisted_paks"].([]interface{}); ok {
		resp.Paks = make([]Pak, 0, len(raw))
		for _, v := range raw {
			m := v.(map[string]interface{})
			resp.Paks = append(resp.Paks, Pak{
				Name: m["name"].(string),
				Size: int64(m["size"].(float64)),
			})
		}
	}

	if fn, ok := config["fortnite"].(map[string]interface{}); ok {
		if v, ok := fn["force_iris"].(bool); ok {
			resp.Fortnite.PatchIris = v
		}
		if v, ok := fn["enable_arena_ui"].(bool); ok {
			resp.Fortnite.EnableArenaHUD = v
		}
		if v, ok := fn["patch_corrupt_data"].(bool); ok {
			resp.Fortnite.PatchCorruptData = v
		}
		if v, ok := fn["patch_iostore"].(bool); ok {
			resp.Fortnite.PatchIOStore = v
		}
		if v, ok := fn["patch_vivox"].(bool); ok {
			resp.Fortnite.PatchVivox = v
		}
	}

	c.JSON(200, resp)
}
