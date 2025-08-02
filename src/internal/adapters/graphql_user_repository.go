package adapters

import (
	"encoding/json"
	"fmt"

	"github.com/18xxnotifier/internal/domain/entities"
)

// GraphQLUserRepository implements entities.UserRepository using GraphQL
type GraphQLUserRepository struct {
	graphqlAdapter GraphQLAdapter
}

// NewGraphQLUserRepository creates a new GraphQL user repository
func NewGraphQLUserRepository(graphqlAdapter GraphQLAdapter) *GraphQLUserRepository {
	return &GraphQLUserRepository{
		graphqlAdapter: graphqlAdapter,
	}
}

// GetUser retrieves a user by Discord ID
func (r *GraphQLUserRepository) GetUser(discordID string) (*entities.User, error) {
	query := `
		query GetUser($discordID: String!) {
			queryUser(filter: { discId: { eq: $discordID } }) {
				discId
				eighteenxxUser
				subscribedGames
			}
		}
	`

	variables := map[string]interface{}{
		"discordID": discordID,
	}

	result, err := r.graphqlAdapter.Query(query, variables)
	if err != nil {
		return nil, fmt.Errorf("failed to query user: %w", err)
	}

	var response struct {
		Data struct {
			QueryUser []struct {
				DiscordID          string   `json:"discId"`
				EighteenxxAccounts []string `json:"eighteenxxUser"`
				SubscribedGames    []string `json:"subscribedGames"`
			} `json:"queryUser"`
		} `json:"data"`
	}

	if err := json.Unmarshal(result, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(response.Data.QueryUser) == 0 {
		return nil, fmt.Errorf("user not found: %s", discordID)
	}

	userData := response.Data.QueryUser[0]
	user := &entities.User{
		DiscordID:          userData.DiscordID,
		EighteenxxAccounts: userData.EighteenxxAccounts,
		SubscribedGames:    userData.SubscribedGames,
	}

	return user, nil
}

// SaveUser saves a new user
func (r *GraphQLUserRepository) SaveUser(user *entities.User) error {
	mutation := `
		mutation AddUser($user: AddUserInput!) {
			addUser(input: [$user]) {
				user {
					discId
				}
			}
		}
	`

	input := map[string]interface{}{
		"discId":          user.DiscordID,
		"eighteenxxUser":  user.EighteenxxAccounts,
		"subscribedGames": user.SubscribedGames,
	}

	variables := map[string]interface{}{
		"user": input,
	}

	_, err := r.graphqlAdapter.Mutate(mutation, variables)
	if err != nil {
		return fmt.Errorf("failed to save user: %w", err)
	}

	return nil
}

// UpdateUser updates an existing user
func (r *GraphQLUserRepository) UpdateUser(user *entities.User) error {
	mutation := `
		mutation UpdateUser($user: UpdateUserInput!) {
			updateUser(input: $user) {
				user {
					discId
				}
			}
		}
	`

	input := map[string]interface{}{
		"filter": map[string]interface{}{
			"discId": map[string]interface{}{
				"eq": user.DiscordID,
			},
		},
		"set": map[string]interface{}{
			"eighteenxxUser":  user.EighteenxxAccounts,
			"subscribedGames": user.SubscribedGames,
		},
	}

	variables := map[string]interface{}{
		"user": input,
	}

	_, err := r.graphqlAdapter.Mutate(mutation, variables)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// DeleteUser deletes a user by Discord ID
func (r *GraphQLUserRepository) DeleteUser(discordID string) error {
	mutation := `
		mutation DeleteUser($filter: UserFilter!) {
			deleteUser(filter: $filter) {
				msg
			}
		}
	`

	filter := map[string]interface{}{
		"discId": map[string]interface{}{
			"eq": discordID,
		},
	}

	variables := map[string]interface{}{
		"filter": filter,
	}

	_, err := r.graphqlAdapter.Mutate(mutation, variables)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

// GetAllUsers retrieves all users
func (r *GraphQLUserRepository) GetAllUsers() ([]*entities.User, error) {
	query := `
		query GetAllUsers {
			queryUser {
				discId
				eighteenxxUser
				subscribedGames
			}
		}
	`

	result, err := r.graphqlAdapter.Query(query, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to query users: %w", err)
	}

	var response struct {
		Data struct {
			QueryUser []struct {
				DiscordID          string   `json:"discId"`
				EighteenxxAccounts []string `json:"eighteenxxUser"`
				SubscribedGames    []string `json:"subscribedGames"`
			} `json:"queryUser"`
		} `json:"data"`
	}

	if err := json.Unmarshal(result, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	var users []*entities.User
	for _, userData := range response.Data.QueryUser {
		user := &entities.User{
			DiscordID:          userData.DiscordID,
			EighteenxxAccounts: userData.EighteenxxAccounts,
			SubscribedGames:    userData.SubscribedGames,
		}
		users = append(users, user)
	}

	return users, nil
}

// GetUsersByGame gets all users subscribed to a specific game
func (r *GraphQLUserRepository) GetUsersByGame(gameID string) ([]*entities.User, error) {
	query := `
		query GetUsersByGame($gameID: String!) {
			queryUser(filter: { subscribedGames: { anyofterms: $gameID } }) {
				discId
				eighteenxxUser
				subscribedGames
			}
		}
	`

	variables := map[string]interface{}{
		"gameID": gameID,
	}

	result, err := r.graphqlAdapter.Query(query, variables)
	if err != nil {
		return nil, fmt.Errorf("failed to query users by game: %w", err)
	}

	var response struct {
		Data struct {
			QueryUser []struct {
				DiscordID          string   `json:"discId"`
				EighteenxxAccounts []string `json:"eighteenxxUser"`
				SubscribedGames    []string `json:"subscribedGames"`
			} `json:"queryUser"`
		} `json:"data"`
	}

	if err := json.Unmarshal(result, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	var users []*entities.User
	for _, userData := range response.Data.QueryUser {
		user := &entities.User{
			DiscordID:          userData.DiscordID,
			EighteenxxAccounts: userData.EighteenxxAccounts,
			SubscribedGames:    userData.SubscribedGames,
		}
		users = append(users, user)
	}

	return users, nil
}
