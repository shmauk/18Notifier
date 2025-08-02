package adapters

import (
	"github.com/18xxnotifier/internal/domain/entities"
)

// GraphQLAdapter defines the interface for GraphQL operations
type GraphQLAdapter interface {
	Query(query string, variables map[string]interface{}) ([]byte, error)
	Mutate(mutation string, variables map[string]interface{}) ([]byte, error)
}

// DiscordCommand represents a command received from Discord
type DiscordCommand struct {
	UserID    string
	ChannelID string
	GuildID   string
	Command   string
	Args      []string
}

// GameDataAdapter defines the interface for fetching game data from 18xx.games API
type GameDataAdapter interface {
	FetchGameData(gameID string) (*entities.Game, error)
}
