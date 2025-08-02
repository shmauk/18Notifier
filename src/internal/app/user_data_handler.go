package app

import (
	"fmt"
	"log"

	"github.com/18xxnotifier/internal/domain/entities"
)

// UserDataHandler handles user data operations and management
type UserDataHandler struct {
	userService *UserService
	userRepo    entities.UserRepository
}

// NewUserDataHandler creates a new user data handler
func NewUserDataHandler(userService *UserService, userRepo entities.UserRepository) *UserDataHandler {
	return &UserDataHandler{
		userService: userService,
		userRepo:    userRepo,
	}
}

// RegisterUser registers a new Discord user
func (h *UserDataHandler) RegisterUser(discordID string) error {
	log.Printf("Registering new user: %s", discordID)

	// Check if user already exists
	existingUser, err := h.userRepo.GetUser(discordID)
	if err == nil && existingUser != nil {
		log.Printf("User %s already registered", discordID)
		return nil // User already exists
	}

	return h.userService.RegisterUser(discordID)
}

// LinkUserToAccount links a Discord user to an 18xx account
func (h *UserDataHandler) LinkUserToAccount(discordID string, eighteenxxAccount string) error {
	log.Printf("Linking user %s to 18xx account: %s", discordID, eighteenxxAccount)

	// Validate account name (basic validation)
	if eighteenxxAccount == "" {
		return fmt.Errorf("18xx account name cannot be empty")
	}

	// Check if user exists, if not register them
	user, err := h.userRepo.GetUser(discordID)
	if err != nil || user == nil {
		log.Printf("User %s not found, registering first", discordID)
		if err := h.RegisterUser(discordID); err != nil {
			return fmt.Errorf("failed to register user: %w", err)
		}
	}

	return h.userService.LinkUserToAccount(discordID, eighteenxxAccount)
}

// UnlinkUserFromAccount removes the link between a Discord user and an 18xx account
func (h *UserDataHandler) UnlinkUserFromAccount(discordID string, eighteenxxAccount string) error {
	log.Printf("Unlinking user %s from 18xx account: %s", discordID, eighteenxxAccount)

	user, err := h.userRepo.GetUser(discordID)
	if err != nil || user == nil {
		return fmt.Errorf("user %s not found", discordID)
	}

	// Check if account is linked
	accountLinked := false
	for _, account := range user.EighteenxxAccounts {
		if account == eighteenxxAccount {
			accountLinked = true
			break
		}
	}

	if !accountLinked {
		return fmt.Errorf("user %s is not linked to account %s", discordID, eighteenxxAccount)
	}

	return h.userService.UnlinkUserFromAccount(discordID, eighteenxxAccount)
}

// SubscribeUserToGame subscribes a user to notifications for a specific game
func (h *UserDataHandler) SubscribeUserToGame(discordID string, gameID string) error {
	log.Printf("Subscribing user %s to game: %s", discordID, gameID)

	// Check if user exists, if not register them
	user, err := h.userRepo.GetUser(discordID)
	if err != nil || user == nil {
		log.Printf("User %s not found, registering first", discordID)
		if err := h.RegisterUser(discordID); err != nil {
			return fmt.Errorf("failed to register user: %w", err)
		}
	}

	return h.userService.SubscribeUserToGame(discordID, gameID)
}

// UnsubscribeUserFromGame unsubscribes a user from notifications for a specific game
func (h *UserDataHandler) UnsubscribeUserFromGame(discordID string, gameID string) error {
	log.Printf("Unsubscribing user %s from game: %s", discordID, gameID)

	user, err := h.userRepo.GetUser(discordID)
	if err != nil || user == nil {
		return fmt.Errorf("user %s not found", discordID)
	}

	// Check if user is subscribed
	subscribed := false
	for _, game := range user.SubscribedGames {
		if game == gameID {
			subscribed = true
			break
		}
	}

	if !subscribed {
		return fmt.Errorf("user %s is not subscribed to game %s", discordID, gameID)
	}

	return h.userService.UnsubscribeUserFromGame(discordID, gameID)
}

// GetUser retrieves a user by Discord ID
func (h *UserDataHandler) GetUser(discordID string) (*entities.User, error) {
	return h.userRepo.GetUser(discordID)
}

// GetUsersByGame gets all users subscribed to a specific game
func (h *UserDataHandler) GetUsersByGame(gameID string) ([]*entities.User, error) {
	return h.userRepo.GetUsersByGame(gameID)
}

// GetAllUsers gets all registered users
func (h *UserDataHandler) GetAllUsers() ([]*entities.User, error) {
	return h.userRepo.GetAllUsers()
}

// GetUserAccounts gets all 18xx accounts linked to a Discord user
func (h *UserDataHandler) GetUserAccounts(discordID string) ([]string, error) {
	user, err := h.userRepo.GetUser(discordID)
	if err != nil {
		return nil, err
	}

	return user.EighteenxxAccounts, nil
}

// GetUserSubscriptions gets all games a user is subscribed to
func (h *UserDataHandler) GetUserSubscriptions(discordID string) ([]string, error) {
	user, err := h.userRepo.GetUser(discordID)
	if err != nil {
		return nil, err
	}

	return user.SubscribedGames, nil
}

// RemoveUser removes a user and all their data
func (h *UserDataHandler) RemoveUser(discordID string) error {
	log.Printf("Removing user: %s", discordID)

	user, err := h.userRepo.GetUser(discordID)
	if err != nil || user == nil {
		return fmt.Errorf("user %s not found", discordID)
	}

	return h.userRepo.DeleteUser(discordID)
}

// UpdateUser updates user information
func (h *UserDataHandler) UpdateUser(user *entities.User) error {
	log.Printf("Updating user: %s", user.DiscordID)
	return h.userRepo.UpdateUser(user)
}
