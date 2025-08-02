package app

import (
	"log"
	"time"

	"github.com/18xxnotifier/internal/domain/entities"
)

// GameDataHandler handles the game data fetching and change detection logic
type GameDataHandler struct {
	gameService *GameService
	interval    time.Duration
	stopChan    chan struct{}
}

// NewGameDataHandler creates a new game data handler
func NewGameDataHandler(gameService *GameService, interval time.Duration) *GameDataHandler {
	return &GameDataHandler{
		gameService: gameService,
		interval:    interval,
		stopChan:    make(chan struct{}),
	}
}

// Start begins the game data monitoring loop
func (h *GameDataHandler) Start() {
	log.Printf("Starting game data handler with %v interval", h.interval)

	ticker := time.NewTicker(h.interval)
	defer ticker.Stop()

	// Run initial check
	h.checkAllGames()

	for {
		select {
		case <-ticker.C:
			h.checkAllGames()
		case <-h.stopChan:
			log.Println("Game data handler stopped")
			return
		}
	}
}

// Stop stops the game data handler
func (h *GameDataHandler) Stop() {
	close(h.stopChan)
}

// checkAllGames checks all tracked games for updates
func (h *GameDataHandler) checkAllGames() {
	log.Println("Checking all tracked games for updates...")

	games, err := h.gameService.GetActiveGames()
	if err != nil {
		log.Printf("Error getting active games: %v", err)
		return
	}

	if len(games) == 0 {
		log.Println("No active games to check")
		return
	}

	log.Printf("Checking %d active games", len(games))

	for _, game := range games {
		h.checkGameForChanges(game.ID)
	}
}

// checkGameForChanges checks a specific game for changes
func (h *GameDataHandler) checkGameForChanges(gameID string) {
	log.Printf("Checking game %s for changes", gameID)

	changes, err := h.gameService.UpdateGameData(gameID)
	if err != nil {
		log.Printf("Error updating game %s: %v", gameID, err)
		return
	}

	if len(changes) > 0 {
		log.Printf("Changes detected in game %s: %d changes", gameID, len(changes))
		for _, change := range changes {
			if change != nil {
				h.handleGameChange(change)
			}
		}
	} else {
		log.Printf("No changes detected in game %s", gameID)
	}
}

// handleGameChange processes a detected game change
func (h *GameDataHandler) handleGameChange(change *entities.GameChange) {
	switch change.ChangeType {
	case "player_change":
		log.Printf("Player change in game %s: %s -> %s",
			change.GameID, change.OldValue, change.NewValue)
	case "game_end":
		log.Printf("Game %s has ended", change.GameID)
		// Remove finished game from tracking
		if err := h.gameService.StopTrackingGame(change.GameID); err != nil {
			log.Printf("Error removing finished game %s: %v", change.GameID, err)
		}
	default:
		log.Printf("Unknown change type in game %s: %s", change.GameID, change.ChangeType)
	}
}

// AddGameToTracking adds a game to the tracking list
func (h *GameDataHandler) AddGameToTracking(gameID string) error {
	log.Printf("Adding game %s to tracking", gameID)
	return h.gameService.TrackGame(gameID)
}

// RemoveGameFromTracking removes a game from the tracking list
func (h *GameDataHandler) RemoveGameFromTracking(gameID string) error {
	log.Printf("Removing game %s from tracking", gameID)
	return h.gameService.StopTrackingGame(gameID)
}

// GetTrackedGames returns all currently tracked games
func (h *GameDataHandler) GetTrackedGames() ([]*entities.Game, error) {
	return h.gameService.GetActiveGames()
}
