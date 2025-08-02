package entities

import "time"

// Game represents a game of 18xx
type Game struct {
	ID           string    `json:"id"`
	Players      []string  `json:"players"`
	ActivePlayer string    `json:"activePlayer"`
	Finished     bool      `json:"finished"`
	LastUpdated  time.Time `json:"lastUpdated"`
}

// IsActive returns true if the game is currently active
func (g *Game) IsActive() bool {
	return !g.Finished && g.ActivePlayer != ""
}

// IsFinished returns true if the game is finished
func (g *Game) IsFinished() bool {
	return g.Finished
}

// GetActivePlayer returns the currently active player
func (g *Game) GetActivePlayer() string {
	return g.ActivePlayer
}

// GetPlayers returns all players in the game
func (g *Game) GetPlayers() []string {
	return g.Players
}

// GetLastUpdated returns the last update time
func (g *Game) GetLastUpdated() time.Time {
	return g.LastUpdated
}

// GameChange represents a change in game state
type GameChange struct {
	GameID     string    `json:"gameId"`
	ChangeType string    `json:"changeType"` // "player_change", "game_end"
	OldValue   string    `json:"oldValue,omitempty"`
	NewValue   string    `json:"newValue,omitempty"`
	Timestamp  time.Time `json:"timestamp"`
}

// GameRepository defines the interface for game data operations
type GameRepository interface {
	GetGame(id string) (*Game, error)
	SaveGame(game *Game) error
	UpdateGame(game *Game) error
	DeleteGame(id string) error
	GetAllGames() ([]*Game, error)
	GetActiveGames() ([]*Game, error)
}
