package adapters

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/18xxnotifier/internal/domain/entities"
)

// EighteenxxAPIAdapter implements GameDataAdapter for the 18xx.games API
type EighteenxxAPIAdapter struct {
	baseURL    string
	httpClient *http.Client
}

// EighteenxxGameData represents the structure of game data from 18xx.games API
type EighteenxxGameData struct {
	ID          int                      `json:"id"`
	Description string                   `json:"description"`
	User        map[string]interface{}   `json:"user"`
	Players     []map[string]interface{} `json:"players"`
	MinPlayers  int                      `json:"min_players"`
	MaxPlayers  int                      `json:"max_players"`
	Title       string                   `json:"title"`
	Settings    map[string]interface{}   `json:"settings"`
	Status      string                   `json:"status"`
	Turn        int                      `json:"turn"`
	Round       string                   `json:"round"`
	Acting      []int                    `json:"acting"`
	Result      map[string]interface{}   `json:"result"`
	Actions     []interface{}            `json:"actions"`
	Loaded      bool                     `json:"loaded"`
	CreatedAt   int64                    `json:"created_at"`
	UpdatedAt   int64                    `json:"updated_at"`
	FinishedAt  *int64                   `json:"finished_at"`
}

// NewEighteenxxAPIAdapter creates a new 18xx API adapter
func NewEighteenxxAPIAdapter() *EighteenxxAPIAdapter {
	return &EighteenxxAPIAdapter{
		baseURL: "https://18xx.games/api/game",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// FetchGameData fetches game data from the 18xx.games API
func (a *EighteenxxAPIAdapter) FetchGameData(gameID string) (*entities.Game, error) {
	url := fmt.Sprintf("%s/%s", a.baseURL, gameID)

	resp, err := a.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch game data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var apiData EighteenxxGameData
	if err := json.Unmarshal(body, &apiData); err != nil {
		return nil, fmt.Errorf("failed to parse API response: %w", err)
	}

	// Convert API data to domain entity
	// Extract player names from the players array
	var playerNames []string
	for _, player := range apiData.Players {
		if name, ok := player["name"].(string); ok {
			playerNames = append(playerNames, name)
		}
	}

	// Determine if game is finished based on status and finished_at
	isFinished := apiData.Status == "finished" || apiData.FinishedAt != nil

	// Get all active players from acting array
	var activePlayers []string
	for _, actingID := range apiData.Acting {
		// Find the player with the acting ID
		for _, player := range apiData.Players {
			if playerID, ok := player["id"].(float64); ok && int(playerID) == actingID {
				if name, ok := player["name"].(string); ok {
					activePlayers = append(activePlayers, name)
					break
				}
			}
		}
	}

	game := &entities.Game{
		ID:            fmt.Sprintf("%d", apiData.ID),
		Players:       playerNames,
		ActivePlayers: activePlayers,
		Finished:      isFinished,
		LastUpdated:   time.Unix(apiData.UpdatedAt, 0),
	}

	return game, nil
}
