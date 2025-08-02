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
	ID           string   `json:"id"`
	Players      []string `json:"players"`
	ActivePlayer string   `json:"active_player"`
	Finished     bool     `json:"finished"`
	LastUpdated  string   `json:"last_updated"`
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
	game := &entities.Game{
		ID:           apiData.ID,
		Players:      apiData.Players,
		ActivePlayer: apiData.ActivePlayer,
		Finished:     apiData.Finished,
		LastUpdated:  time.Now(), // We'll use current time since API doesn't provide it
	}

	return game, nil
}
