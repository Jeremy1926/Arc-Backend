package messages

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/Arc-Services/Arc/services/anticheat/main/classes"
	"github.com/Arc-Services/Arc/services/anticheat/util"
	"github.com/Arc-Services/Arc/shared/cast"
	database "github.com/Arc-Services/Arc/shared/database/classes/anticheat"
	main "github.com/Arc-Services/Arc/shared/database/classes/main"
	"github.com/Arc-Services/Arc/shared/middleware"
)

type DiscordEmbedFooter struct {
	Text    string `json:"text,omitempty"`
	IconUrl string `json:"icon_url,omitempty"`
}

type DiscordEmbedField struct {
	Name   string `json:"name,omitempty"`
	Value  string `json:"value,omitempty"`
	Inline bool   `json:"inline,omitempty"`
}

type DiscordEmbed struct {
	Title       string              `json:"title,omitempty"`
	Url         string              `json:"url,omitempty"`
	Description string              `json:"description,omitempty"`
	Color       int                 `json:"color,omitempty"`
	Footer      DiscordEmbedFooter  `json:"footer"`
	Fields      []DiscordEmbedField `json:"fields,omitempty"`
	Timestamp   time.Time           `json:"timestamp"`
}

type DiscordWebhookPayload struct {
	Username  string         `json:"username,omitempty"`
	AvatarUrl string         `json:"avatar_url,omitempty"`
	Content   string         `json:"content,omitempty"`
	Embeds    []DiscordEmbed `json:"embeds,omitempty"`
}

func Detection(client *classes.PublicClient, data map[string]interface{}) error {
	if !client.HasCompletedChallenge {
		util.RemoveClient(client)
		return errors.New("validation failed")
	}
	info := "No additional information."
	if infoRaw, ok := data["info"].(string); ok {
		info = infoRaw
	}

	codeInt := int(data["code"].(float64))
	code := classes.DetectionCode(codeInt)
	banMap := map[int]bool{
		0:  false, // UnsignedDLLLoaded
		3:  false, // CreatedOverlay
		6:  false, // ManualMappedDLL
		10: false, // BadThreadCreate
		12: false, // SuspiciousPakLoad
		21: false,
		23: false,
	}
	ban := true
	if banVal, ok := banMap[codeInt]; ok {
		ban = banVal
	}
	codeAsString := code.String()

	dtc := &database.Detection{
		AccountID: client.AccountID,
		Type:      codeAsString,
		Info:      info,
		Ban:       ban,
	}

	id := middleware.MustGetClientID(client.Context)
	db := middleware.MustGetClientDB(client.Context)
	if db == nil {
		util.RemoveClient(client)
		return errors.New("client database is nil")
	}

	if client.Configuration != nil {
		discord, ok := client.Configuration["discord"].(map[string]interface{})
		if ok {
			webhook, ok := discord["webhook"].(string)
			if ok && webhook != "" {
				body, _ := json.Marshal(DiscordWebhookPayload{
					Embeds: []DiscordEmbed{
						{
							Title: "Received detection report",
							Color: 16711680,
							Fields: []DiscordEmbedField{
								{
									Name:  "Display name",
									Value: client.DisplayName,
								},
								{
									Name:  "Account ID",
									Value: client.AccountID,
								},
								{
									Name:  "Detection type",
									Value: codeAsString,
								},
								{
									Name:  "Additional info",
									Value: info,
								},
								{
									Name:  "Player banned",
									Value: fmt.Sprintf("%v", ban),
								},
							},
							Footer: DiscordEmbedFooter{
								Text: "Arc | Detections",
							},
							Timestamp: time.Now(),
						},
					},
				})

				http.Post(webhook, "application/json", bytes.NewBuffer(body))
			}
		}
	}

	db.SetJSON(fmt.Sprintf("detection:%s:%d", client.AccountID, rand.Int()), dtc)

	if ban {
		api, ok := client.Configuration["api"].(map[string]interface{})
		if ok {
			backend, ok := api["outgoing"].(string)
			if ok && backend != "" {
				go util.SendJSON(client.Configuration, fmt.Sprintf("%s/outgoing/v1/anticheat/private/ban", backend), map[string]interface{}{
					"id": client.AccountID,
				})
			}
		}

		res, err := account_db.GetFor(id, client.AccountID)
		if err != nil {
			return err
		}

		m, ok := res.Value.(map[string]any)
		if ok {
			account := &main.Account{
				DisplayName: cast.Str(m["display_name"]),
				Country:     cast.Str(m["country"]),
				Active:      false,
			}

			account_db.ClientID = id
			account_db.Set(client.AccountID, account)
			account_db.Set(fmt.Sprintf("account:%s", client.AccountID), account)
			account_db.Set(fmt.Sprintf("account:%s", strings.ToLower(account.DisplayName)), account)
		}

		util.RemoveClient(client)
	}

	return nil
}
