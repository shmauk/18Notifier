package entities

// Guild represents a Discord server/guild
type Guild struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Channels []string `json:"channels"` // Channel IDs
}

// HasChannel returns true if the guild has the specified channel
func (g *Guild) HasChannel(channelID string) bool {
	for _, channel := range g.Channels {
		if channel == channelID {
			return true
		}
	}
	return false
}

// AddChannel adds a channel to the guild if it doesn't already exist
func (g *Guild) AddChannel(channelID string) {
	if !g.HasChannel(channelID) {
		g.Channels = append(g.Channels, channelID)
	}
}

// RemoveChannel removes a channel from the guild
func (g *Guild) RemoveChannel(channelID string) {
	for i, channel := range g.Channels {
		if channel == channelID {
			g.Channels = append(g.Channels[:i], g.Channels[i+1:]...)
			break
		}
	}
}

// GetChannels returns all channels in this guild
func (g *Guild) GetChannels() []string {
	return g.Channels
}

// GetName returns the guild name
func (g *Guild) GetName() string {
	return g.Name
}

// GetID returns the guild ID
func (g *Guild) GetID() string {
	return g.ID
}

// Channel represents a Discord channel
type Channel struct {
	ID      string   `json:"id"`
	GuildID string   `json:"guildId"`
	Name    string   `json:"name"`
	Games   []string `json:"games"` // Game IDs being tracked
}

// HasGame returns true if the channel is tracking the specified game
func (c *Channel) HasGame(gameID string) bool {
	for _, game := range c.Games {
		if game == gameID {
			return true
		}
	}
	return false
}

// AddGame adds a game to the channel's tracking list if it doesn't already exist
func (c *Channel) AddGame(gameID string) {
	if !c.HasGame(gameID) {
		c.Games = append(c.Games, gameID)
	}
}

// RemoveGame removes a game from the channel's tracking list
func (c *Channel) RemoveGame(gameID string) {
	for i, game := range c.Games {
		if game == gameID {
			c.Games = append(c.Games[:i], c.Games[i+1:]...)
			break
		}
	}
}

// GetGames returns all games being tracked by this channel
func (c *Channel) GetGames() []string {
	return c.Games
}

// GetName returns the channel name
func (c *Channel) GetName() string {
	return c.Name
}

// GetGuildID returns the guild ID this channel belongs to
func (c *Channel) GetGuildID() string {
	return c.GuildID
}

// GuildRepository defines the interface for guild data operations
type GuildRepository interface {
	GetGuild(id string) (*Guild, error)
	SaveGuild(guild *Guild) error
	UpdateGuild(guild *Guild) error
	DeleteGuild(id string) error
	GetAllGuilds() ([]*Guild, error)
}

// ChannelRepository defines the interface for channel data operations
type ChannelRepository interface {
	GetChannel(id string) (*Channel, error)
	SaveChannel(channel *Channel) error
	UpdateChannel(channel *Channel) error
	DeleteChannel(id string) error
	GetAllChannels() ([]*Channel, error)
	GetChannelsByGuild(guildID string) ([]*Channel, error)
	GetChannelsByGame(gameID string) ([]*Channel, error)
}
