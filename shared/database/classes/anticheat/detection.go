package database

type Detection struct {
	AccountID string `json:"account_id"`
	Type      string `json:"type"`
	Info      string `json:"info"`
	Ban       bool   `json:"ban"`
}
