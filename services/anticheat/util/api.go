package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	anticheat "github.com/Arc-Services/Arc/services/anticheat/main"
)

func SendJSON(configuration map[string]interface{}, url string, data map[string]interface{}) {
	body, err := json.Marshal(data)
	if err != nil {
		return
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return
	}

	req.Header.Set("User-Agent", fmt.Sprintf("Arc/%v", anticheat.Version))
	req.Header.Set("Content-Type", "application/json")

	api, ok := configuration["api"].(map[string]interface{})
	if ok {
		key, ok := api["key"].(string)
		if ok && key != "" {
			req.Header.Set("X-Arc-Auth", key)
		}
	}

	client := &http.Client{}
	_, _ = client.Do(req)
}
