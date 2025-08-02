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
				activePlayers
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
				ID           string   `json:"id"`
				Players      []string `json:"players"`
				ActivePlayer string   `json:"activePlayers"`
				Finished     bool     `json:"finished"`
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
		ID:           gameData.ID,
		Players:      gameData.Players,
		ActivePlayer: gameData.ActivePlayer,
		Finished:     gameData.Finished,
	}

	return game, nil
}

// SaveGame saves a new game
func (r *GraphQLGameRepository) SaveGame(game *entities.Game) error {
	mutation := `
		mutation AddGame($game: AddGameDataInput!) {
			addGameData(input: [$game]) {
				gameData {
					id
				}
			}
		}
	`

	input := map[string]interface{}{
		"id":            game.ID,
		"players":       game.Players,
		"activePlayers": game.ActivePlayer,
		"finished":      game.Finished,
	}

	variables := map[string]interface{}{
		"game": input,
	}

	_, err := r.graphqlAdapter.Mutate(mutation, variables)
	if err != nil {
		return fmt.Errorf("failed to save game: %w", err)
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
			"activePlayers": game.ActivePlayer,
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
				ID           string   `json:"id"`
				Players      []string `json:"players"`
				ActivePlayer string   `json:"activePlayers"`
				Finished     bool     `json:"finished"`
			} `json:"queryGameData"`
		} `json:"data"`
	}

	if err := json.Unmarshal(result, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	var games []*entities.Game
	for _, gameData := range response.Data.QueryGameData {
		game := &entities.Game{
			ID:           gameData.ID,
			Players:      gameData.Players,
			ActivePlayer: gameData.ActivePlayer,
			Finished:     gameData.Finished,
		}
		games = append(games, game)
	}

	return games, nil
}

// GetActiveGames retrieves all active (non-finished) games
func (r *GraphQLGameRepository) GetActiveGames() ([]*entities.Game, error) {
	query := `
		query GetActiveGames {
			queryGameData(filter: { finished: { eq: false } }) {
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
				ID           string   `json:"id"`
				Players      []string `json:"players"`
				ActivePlayer string   `json:"activePlayers"`
				Finished     bool     `json:"finished"`
			} `json:"queryGameData"`
		} `json:"data"`
	}

	if err := json.Unmarshal(result, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	var games []*entities.Game
	for _, gameData := range response.Data.QueryGameData {
		game := &entities.Game{
			ID:           gameData.ID,
			Players:      gameData.Players,
			ActivePlayer: gameData.ActivePlayer,
			Finished:     gameData.Finished,
		}
		games = append(games, game)
	}

	return games, nil
}
