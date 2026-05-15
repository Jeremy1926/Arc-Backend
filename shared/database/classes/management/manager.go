package management

type Manager struct {
	AccountId   string   `json:"account_id"`
	DisplayName string   `json:"display_name"`
	Country     string   `json:"country"`
	Email       string   `json:"email"`
	Password    string   `json:"password"`
	Permissions []string `json:"permissions"`
}
