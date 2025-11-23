package domain

// Team представляет команду пользователей.
type Team struct {
	Name    string `json:"team_name"`
	Members []User `json:"members"`
}
