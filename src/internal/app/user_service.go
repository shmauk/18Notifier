package app

import (
	"fmt"

	"github.com/18xxnotifier/internal/domain/entities"
)

// UserService handles user-related business logic
type UserService struct {
	userRepo entities.UserRepository
}

// NewUserService creates a new UserService instance
func NewUserService(userRepo entities.UserRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

// RegisterUser registers a new user
func (s *UserService) RegisterUser(discordID string) error {
	// Check if user already exists
	existingUser, err := s.userRepo.GetUser(discordID)
	if err != nil {
		return fmt.Errorf("failed to check existing user: %w", err)
	}

	if existingUser != nil {
		return fmt.Errorf("user already exists: %s", discordID)
	}

	// Create new user
	user := &entities.User{
		DiscordID:          discordID,
		EighteenxxAccounts: []string{},
		SubscribedGames:    []string{},
	}

	// Save user to repository
	err = s.userRepo.SaveUser(user)
	if err != nil {
		return fmt.Errorf("failed to save user: %w", err)
	}

	return nil
}

// LinkUserToAccount links a Discord user to an 18xx account
func (s *UserService) LinkUserToAccount(discordID, account string) error {
	// Get user
	user, err := s.userRepo.GetUser(discordID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	if user == nil {
		return fmt.Errorf("user not found: %s", discordID)
	}

	// Add account if not already linked
	user.AddAccount(account)

	// Update user in repository
	err = s.userRepo.UpdateUser(user)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// UnlinkUserFromAccount unlinks a Discord user from an 18xx account
func (s *UserService) UnlinkUserFromAccount(discordID, account string) error {
	// Get user
	user, err := s.userRepo.GetUser(discordID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	if user == nil {
		return fmt.Errorf("user not found: %s", discordID)
	}

	// Remove account
	user.RemoveAccount(account)

	// Update user in repository
	err = s.userRepo.UpdateUser(user)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// SubscribeUserToGame subscribes a user to game notifications
func (s *UserService) SubscribeUserToGame(discordID, gameID string) error {
	// Get user
	user, err := s.userRepo.GetUser(discordID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	if user == nil {
		return fmt.Errorf("user not found: %s", discordID)
	}

	// Subscribe to game
	user.SubscribeToGame(gameID)

	// Update user in repository
	err = s.userRepo.UpdateUser(user)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// UnsubscribeUserFromGame unsubscribes a user from game notifications
func (s *UserService) UnsubscribeUserFromGame(discordID, gameID string) error {
	// Get user
	user, err := s.userRepo.GetUser(discordID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	if user == nil {
		return fmt.Errorf("user not found: %s", discordID)
	}

	// Unsubscribe from game
	user.UnsubscribeFromGame(gameID)

	// Update user in repository
	err = s.userRepo.UpdateUser(user)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// GetUser retrieves a user by Discord ID
func (s *UserService) GetUser(discordID string) (*entities.User, error) {
	return s.userRepo.GetUser(discordID)
}

// GetUsersByGame returns all users subscribed to a specific game
func (s *UserService) GetUsersByGame(gameID string) ([]*entities.User, error) {
	return s.userRepo.GetUsersByGame(gameID)
}

// GetAllUsers returns all users
func (s *UserService) GetAllUsers() ([]*entities.User, error) {
	return s.userRepo.GetAllUsers()
}
