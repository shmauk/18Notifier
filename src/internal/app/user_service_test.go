package app

import (
	"testing"

	"github.com/18xxnotifier/internal/domain/entities"
)

// MockUserRepository implements entities.UserRepository for testing
type MockUserRepository struct {
	users map[string]*entities.User
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		users: make(map[string]*entities.User),
	}
}

func (m *MockUserRepository) GetUser(discordID string) (*entities.User, error) {
	if user, exists := m.users[discordID]; exists {
		return user, nil
	}
	return nil, nil
}

func (m *MockUserRepository) SaveUser(user *entities.User) error {
	m.users[user.DiscordID] = user
	return nil
}

func (m *MockUserRepository) UpdateUser(user *entities.User) error {
	m.users[user.DiscordID] = user
	return nil
}

func (m *MockUserRepository) DeleteUser(discordID string) error {
	delete(m.users, discordID)
	return nil
}

func (m *MockUserRepository) GetAllUsers() ([]*entities.User, error) {
	var users []*entities.User
	for _, user := range m.users {
		users = append(users, user)
	}
	return users, nil
}

func (m *MockUserRepository) GetUsersByGame(gameID string) ([]*entities.User, error) {
	var users []*entities.User
	for _, user := range m.users {
		if user.IsSubscribedToGame(gameID) {
			users = append(users, user)
		}
	}
	return users, nil
}

func TestUserService_RegisterUser(t *testing.T) {
	repo := NewMockUserRepository()
	service := NewUserService(repo)

	// Test registering a new user
	discordID := "123456789"
	err := service.RegisterUser(discordID)
	if err != nil {
		t.Errorf("UserService.RegisterUser() error = %v", err)
	}

	// Verify user was created
	user, err := repo.GetUser(discordID)
	if err != nil {
		t.Errorf("Failed to get registered user: %v", err)
	}
	if user == nil {
		t.Error("Registered user not found in repository")
	}
	if user.DiscordID != discordID {
		t.Errorf("Expected Discord ID '%s', got '%s'", discordID, user.DiscordID)
	}
}

func TestUserService_LinkUserToAccount(t *testing.T) {
	repo := NewMockUserRepository()
	service := NewUserService(repo)

	// Register a user first
	discordID := "123456789"
	service.RegisterUser(discordID)

	// Test linking account
	account := "player1"
	err := service.LinkUserToAccount(discordID, account)
	if err != nil {
		t.Errorf("UserService.LinkUserToAccount() error = %v", err)
	}

	// Verify account was linked
	user, err := repo.GetUser(discordID)
	if err != nil {
		t.Errorf("Failed to get user: %v", err)
	}
	if !user.HasAccount(account) {
		t.Errorf("Expected user to have account '%s'", account)
	}
}

func TestUserService_UnlinkUserFromAccount(t *testing.T) {
	repo := NewMockUserRepository()
	service := NewUserService(repo)

	// Register and link a user
	discordID := "123456789"
	account := "player1"
	service.RegisterUser(discordID)
	service.LinkUserToAccount(discordID, account)

	// Test unlinking account
	err := service.UnlinkUserFromAccount(discordID, account)
	if err != nil {
		t.Errorf("UserService.UnlinkUserFromAccount() error = %v", err)
	}

	// Verify account was unlinked
	user, err := repo.GetUser(discordID)
	if err != nil {
		t.Errorf("Failed to get user: %v", err)
	}
	if user.HasAccount(account) {
		t.Errorf("Expected user to not have account '%s'", account)
	}
}

func TestUserService_SubscribeUserToGame(t *testing.T) {
	repo := NewMockUserRepository()
	service := NewUserService(repo)

	// Register a user first
	discordID := "123456789"
	service.RegisterUser(discordID)

	// Test subscribing to game
	gameID := "game123"
	err := service.SubscribeUserToGame(discordID, gameID)
	if err != nil {
		t.Errorf("UserService.SubscribeUserToGame() error = %v", err)
	}

	// Verify subscription was added
	user, err := repo.GetUser(discordID)
	if err != nil {
		t.Errorf("Failed to get user: %v", err)
	}
	if !user.IsSubscribedToGame(gameID) {
		t.Errorf("Expected user to be subscribed to game '%s'", gameID)
	}
}

func TestUserService_UnsubscribeUserFromGame(t *testing.T) {
	repo := NewMockUserRepository()
	service := NewUserService(repo)

	// Register and subscribe a user
	discordID := "123456789"
	gameID := "game123"
	service.RegisterUser(discordID)
	service.SubscribeUserToGame(discordID, gameID)

	// Test unsubscribing from game
	err := service.UnsubscribeUserFromGame(discordID, gameID)
	if err != nil {
		t.Errorf("UserService.UnsubscribeUserFromGame() error = %v", err)
	}

	// Verify subscription was removed
	user, err := repo.GetUser(discordID)
	if err != nil {
		t.Errorf("Failed to get user: %v", err)
	}
	if user.IsSubscribedToGame(gameID) {
		t.Errorf("Expected user to not be subscribed to game '%s'", gameID)
	}
}

func TestUserService_GetUser(t *testing.T) {
	repo := NewMockUserRepository()
	service := NewUserService(repo)

	// Add test user
	user := &entities.User{
		DiscordID:          "123456789",
		EighteenxxAccounts: []string{"player1"},
		SubscribedGames:    []string{"game1"},
	}
	repo.SaveUser(user)

	// Test getting user
	retrievedUser, err := service.GetUser("123456789")
	if err != nil {
		t.Errorf("UserService.GetUser() error = %v", err)
	}

	if retrievedUser == nil {
		t.Error("Expected user to be retrieved, got nil")
	}

	if retrievedUser.DiscordID != "123456789" {
		t.Errorf("Expected Discord ID '123456789', got '%s'", retrievedUser.DiscordID)
	}
}

func TestUserService_GetUsersByGame(t *testing.T) {
	repo := NewMockUserRepository()
	service := NewUserService(repo)

	// Add test users
	user1 := &entities.User{
		DiscordID:          "user1",
		EighteenxxAccounts: []string{"player1"},
		SubscribedGames:    []string{"game1", "game2"},
	}
	user2 := &entities.User{
		DiscordID:          "user2",
		EighteenxxAccounts: []string{"player2"},
		SubscribedGames:    []string{"game1"},
	}
	user3 := &entities.User{
		DiscordID:          "user3",
		EighteenxxAccounts: []string{"player3"},
		SubscribedGames:    []string{"game3"},
	}

	repo.SaveUser(user1)
	repo.SaveUser(user2)
	repo.SaveUser(user3)

	// Test getting users by game
	users, err := service.GetUsersByGame("game1")
	if err != nil {
		t.Errorf("UserService.GetUsersByGame() error = %v", err)
	}

	if len(users) != 2 {
		t.Errorf("Expected 2 users subscribed to game1, got %d", len(users))
	}

	// Verify correct users were returned
	foundUser1 := false
	foundUser2 := false
	for _, user := range users {
		if user.DiscordID == "user1" {
			foundUser1 = true
		}
		if user.DiscordID == "user2" {
			foundUser2 = true
		}
	}

	if !foundUser1 || !foundUser2 {
		t.Error("Expected users 'user1' and 'user2' to be found")
	}
}

func TestUserService_GetAllUsers(t *testing.T) {
	repo := NewMockUserRepository()
	service := NewUserService(repo)

	// Add test users
	user1 := &entities.User{DiscordID: "user1", EighteenxxAccounts: []string{"player1"}}
	user2 := &entities.User{DiscordID: "user2", EighteenxxAccounts: []string{"player2"}}
	repo.SaveUser(user1)
	repo.SaveUser(user2)

	// Test getting all users
	users, err := service.GetAllUsers()
	if err != nil {
		t.Errorf("UserService.GetAllUsers() error = %v", err)
	}

	if len(users) != 2 {
		t.Errorf("Expected 2 users, got %d", len(users))
	}
}
