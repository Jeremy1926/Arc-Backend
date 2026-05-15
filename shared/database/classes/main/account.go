package database

type Account struct {
	DisplayName string `json:"display_name"`
	Country     string `json:"country"`
	Active      bool   `json:"active"`
}
