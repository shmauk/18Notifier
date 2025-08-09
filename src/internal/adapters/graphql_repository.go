package adapters

import (
	"encoding/json"
	"fmt"

	"github.com/18xxnotifier/internal/domain/entities"
)

// GraphQLGameRepository implements entities.GameRepository using GraphQL
type GraphQLGameRepository struct {
	graphqlAdapter GraphQLAdapter
}

// NewGraphQLGameRepository creates a new GraphQL game repository
func NewGraphQLGameRepository(graphqlAdapter GraphQLAdapter) *GraphQLGameRepository {
	return &GraphQLGameRepository{
		graphqlAdapter: graphqlAdapter,
	}
}

// GetGame retrieves a game by ID
func (r *GraphQLGameRepository) GetGame(id string) (*entities.Game, error) {
	query := `
		query GetGame($id: String!) {
			queryGameData(filter: { id: { eq: $id } }) {
				id
				players
				activePlayer
				finished
			}
		}
	`

	variables := map[string]interface{}{
		"id": id,
	}

	result, err := r.graphqlAdapter.Query(query, variables)
	if err != nil {
		return nil, fmt.Errorf("failed to query game: %w", err)
	}

	var response struct {
		Data struct {
			QueryGameData []struct {
				ID            string   `json:"id"`
				Players       []string `json:"players"`
				ActivePlayers []string `json:"activePlayers"`
				Finished      bool     `json:"finished"`
			} `json:"queryGameData"`
		} `json:"data"`
	}

	if err := json.Unmarshal(result, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(response.Data.QueryGameData) == 0 {
		return nil, fmt.Errorf("game not found: %s", id)
	}

	gameData := response.Data.QueryGameData[0]
	game := &entities.Game{
		ID:            gameData.ID,
		Players:       gameData.Players,
		ActivePlayers: gameData.ActivePlayers,
		Finished:      gameData.Finished,
	}

	return game, nil
}

// SaveGame saves a new game
func (r *GraphQLGameRepository) SaveGame(game *entities.Game, channelID string, guildID string) error {
	// First, ensure the guild exists
	guildMutation := `
		mutation AddGuild($guild: AddGuildInput!) {
			addGuild(input: [$guild]) {
				guild {
					id
				}
			}
		}
	`

	guildInput := map[string]interface{}{
		"id": guildID,
	}

	guildVariables := map[string]interface{}{
		"guild": guildInput,
	}

	_, err := r.graphqlAdapter.Mutate(guildMutation, guildVariables)
	if err != nil {
		// Guild might already exist, continue
	}

	// Then, ensure the channel exists
	channelMutation := `
		mutation AddChannel($channel: AddChannelInput!) {
			addChannel(input: [$channel]) {
				channel {
					id
				}
			}
		}
	`

	channelInput := map[string]interface{}{
		"id": channelID,
		"guild": map[string]interface{}{
			"id": guildID,
		},
	}

	channelVariables := map[string]interface{}{
		"channel": channelInput,
	}

	_, err = r.graphqlAdapter.Mutate(channelMutation, channelVariables)
	if err != nil {
		// Channel might already exist, continue
	}

	// Check if GameData already exists
	existingGame, err := r.GetGame(game.ID)
	if err != nil || existingGame == nil {
		// GameData doesn't exist, create it with nested GameChannelMap
		gameMutation := `
			mutation AddGame($game: AddGameDataInput!) {
				addGameData(input: [$game]) {
					gameData {
						id
						channel {
							game {
								id
							}
							channel {
								id
							}
						}
					}
				}
			}
		`

		gameInput := map[string]interface{}{
			"id":            game.ID,
			"players":       game.Players,
			"activePlayers": game.ActivePlayers,
			"finished":      game.Finished,
			"channel": map[string]interface{}{
				"game": map[string]interface{}{
					"id": game.ID,
				},
				"channel": map[string]interface{}{
					"id": channelID,
				},
				"users": []map[string]interface{}{}, // Empty users array for now
			},
		}

		gameVariables := map[string]interface{}{
			"game": gameInput,
		}

		_, err = r.graphqlAdapter.Mutate(gameMutation, gameVariables)
		if err != nil {
			return fmt.Errorf("failed to create new game: %w", err)
		}
	} else {
		// GameData exists, create only a new GameChannelMap
		channelMapMutation := `
			mutation AddGameChannelMap($channelMap: AddGameChannelMapInput!) {
				addGameChannelMap(input: [$channelMap]) {
					gameChannelMap {
						game {
							id
						}
						channel {
							id
						}
					}
				}
			}
		`

		channelMapInput := map[string]interface{}{
			"game": map[string]interface{}{
				"id": game.ID,
			},
			"channel": map[string]interface{}{
				"id": channelID,
			},
			"users": []map[string]interface{}{}, // Empty users array for now
		}

		channelMapVariables := map[string]interface{}{
			"channelMap": channelMapInput,
		}

		_, err = r.graphqlAdapter.Mutate(channelMapMutation, channelMapVariables)
		if err != nil {
			return fmt.Errorf("failed to create game channel mapping: %w", err)
		}
	}

	return nil
}

// UpdateGame updates an existing game
func (r *GraphQLGameRepository) UpdateGame(game *entities.Game) error {
	mutation := `
		mutation UpdateGame($game: UpdateGameDataInput!) {
			updateGameData(input: $game) {
				gameData {
					id
				}
			}
		}
	`

	input := map[string]interface{}{
		"filter": map[string]interface{}{
			"id": map[string]interface{}{
				"eq": game.ID,
			},
		},
		"set": map[string]interface{}{
			"players":       game.Players,
			"activePlayers": game.ActivePlayers,
			"finished":      game.Finished,
		},
	}

	variables := map[string]interface{}{
		"game": input,
	}

	_, err := r.graphqlAdapter.Mutate(mutation, variables)
	if err != nil {
		return fmt.Errorf("failed to update game: %w", err)
	}

	return nil
}

// DeleteGame deletes a game by ID
func (r *GraphQLGameRepository) DeleteGame(id string) error {
	mutation := `
		mutation DeleteGame($filter: GameDataFilter!) {
			deleteGameData(filter: $filter) {
				msg
			}
		}
	`

	filter := map[string]interface{}{
		"id": map[string]interface{}{
			"eq": id,
		},
	}

	variables := map[string]interface{}{
		"filter": filter,
	}

	_, err := r.graphqlAdapter.Mutate(mutation, variables)
	if err != nil {
		return fmt.Errorf("failed to delete game: %w", err)
	}

	return nil
}

// GetAllGames retrieves all games
func (r *GraphQLGameRepository) GetAllGames() ([]*entities.Game, error) {
	query := `
		query GetAllGames {
			queryGameData {
				id
				players
				activePlayers
				finished
			}
		}
	`

	result, err := r.graphqlAdapter.Query(query, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to query games: %w", err)
	}

	var response struct {
		Data struct {
			QueryGameData []struct {
				ID            string   `json:"id"`
				Players       []string `json:"players"`
				ActivePlayers []string `json:"activePlayers"`
				Finished      bool     `json:"finished"`
			} `json:"queryGameData"`
		} `json:"data"`
	}

	if err := json.Unmarshal(result, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	var games []*entities.Game
	for _, gameData := range response.Data.QueryGameData {
		game := &entities.Game{
			ID:            gameData.ID,
			Players:       gameData.Players,
			ActivePlayers: gameData.ActivePlayers,
			Finished:      gameData.Finished,
		}
		games = append(games, game)
	}

	return games, nil
}

// GetActiveGames retrieves all active (non-finished) games
func (r *GraphQLGameRepository) GetActiveGames() ([]*entities.Game, error) {
	query := `
		query GetActiveGames {
			queryGameData {
				id
				players
				activePlayers
				finished
			}
		}
	`

	result, err := r.graphqlAdapter.Query(query, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to query active games: %w", err)
	}

	var response struct {
		Data struct {
			QueryGameData []struct {
				ID            string   `json:"id"`
				Players       []string `json:"players"`
				ActivePlayers []string `json:"activePlayers"`
				Finished      bool     `json:"finished"`
			} `json:"queryGameData"`
		} `json:"data"`
	}

	if err := json.Unmarshal(result, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	var games []*entities.Game
	for _, gameData := range response.Data.QueryGameData {
		// Filter for non-finished games in application code
		if !gameData.Finished {
			game := &entities.Game{
				ID:            gameData.ID,
				Players:       gameData.Players,
				ActivePlayers: gameData.ActivePlayers,
				Finished:      gameData.Finished,
			}
			games = append(games, game)
		}
	}

	return games, nil
}
