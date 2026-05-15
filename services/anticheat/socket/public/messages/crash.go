package messages

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/Arc-Services/Arc/services/anticheat/main/classes"
)

func Crash(client *classes.PublicClient, data map[string]interface{}) error {
	if client.Configuration != nil {
		if discord, ok := client.Configuration["discord"].(map[string]interface{}); ok {
			if webhook, ok := discord["webhook"].(string); ok && webhook != "" {
				info, _ := data["info"].(string)
				info = strings.ReplaceAll(info, `\r\n`, "\n")
				info = strings.ReplaceAll(info, `\n`, "\n")
				info = strings.ReplaceAll(info, `\t`, "\t")

				body, _ := json.Marshal(map[string]string{
					"content": fmt.Sprintf(
						"crash: displayname=%s, info=%s",
						client.DisplayName,
						info,
					),
				})
				http.Post(webhook, "application/json", bytes.NewBuffer(body))
			}
		}
	}
	return nil
}
