package database

type Session struct {
	Owner       string   `json:"owner"`
	Permissions []string `json:"permissions"`
	Issuer      string   `json:"issuer"`
	Client      string   `json:"client"`
}
