package entities

// User represents a Discord user with their 18xx account mappings
type User struct {
	DiscordID          string   `json:"discordId"`
	EighteenxxAccounts []string `json:"eighteenxxAccounts"`
	SubscribedGames    []string `json:"subscribedGames"`
}

// HasAccount returns true if the user has the specified 18xx account
func (u *User) HasAccount(account string) bool {
	for _, acc := range u.EighteenxxAccounts {
		if acc == account {
			return true
		}
	}
	return false
}

// IsSubscribedToGame returns true if the user is subscribed to the specified game
func (u *User) IsSubscribedToGame(gameID string) bool {
	for _, game := range u.SubscribedGames {
		if game == gameID {
			return true
		}
	}
	return false
}

// AddAccount adds an 18xx account to the user if it doesn't already exist
func (u *User) AddAccount(account string) {
	if !u.HasAccount(account) {
		u.EighteenxxAccounts = append(u.EighteenxxAccounts, account)
	}
}

// RemoveAccount removes an 18xx account from the user
func (u *User) RemoveAccount(account string) {
	for i, acc := range u.EighteenxxAccounts {
		if acc == account {
			u.EighteenxxAccounts = append(u.EighteenxxAccounts[:i], u.EighteenxxAccounts[i+1:]...)
			break
		}
	}
}

// SubscribeToGame adds a game subscription if it doesn't already exist
func (u *User) SubscribeToGame(gameID string) {
	if !u.IsSubscribedToGame(gameID) {
		u.SubscribedGames = append(u.SubscribedGames, gameID)
	}
}

// UnsubscribeFromGame removes a game subscription
func (u *User) UnsubscribeFromGame(gameID string) {
	for i, game := range u.SubscribedGames {
		if game == gameID {
			u.SubscribedGames = append(u.SubscribedGames[:i], u.SubscribedGames[i+1:]...)
			break
		}
	}
}

// UserRepository defines the interface for user data operations
type UserRepository interface {
	GetUser(discordID string) (*User, error)
	SaveUser(user *User) error
	UpdateUser(user *User) error
	DeleteUser(discordID string) error
	GetAllUsers() ([]*User, error)
	GetUsersByGame(gameID string) ([]*User, error)
}
