package api

type Error struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
	Route   string `json:"route"`
}
