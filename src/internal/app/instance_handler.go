package app

import (
	"log"

	"github.com/18xxnotifier/internal/adapters"
	"github.com/18xxnotifier/internal/domain/entities"
)

// InstanceHandler handles Discord server/guild and channel management
type InstanceHandler struct {
	guildRepo      entities.GuildRepository
	channelRepo    entities.ChannelRepository
	discordAdapter *adapters.DiscordAdapter
}

// NewInstanceHandler creates a new instance handler
func NewInstanceHandler(
	guildRepo entities.GuildRepository,
	channelRepo entities.ChannelRepository,
	discordAdapter *adapters.DiscordAdapter,
) *InstanceHandler {
	return &InstanceHandler{
		guildRepo:      guildRepo,
		channelRepo:    channelRepo,
		discordAdapter: discordAdapter,
	}
}

// HandleDiscordCommand processes Discord commands related to instance management
func (h *InstanceHandler) HandleDiscordCommand(cmd adapters.DiscordCommand) {
	switch cmd.Command {
	case "addserver":
		h.handleAddServer(cmd)
	case "addchannel":
		h.handleAddChannel(cmd)
	case "removeserver":
		h.handleRemoveServer(cmd)
	case "removechannel":
		h.handleRemoveChannel(cmd)
	case "listchannels":
		h.handleListChannels(cmd)
	default:
		// Not an instance management command
		return
	}
}

// handleAddServer adds a new Discord server/guild
func (h *InstanceHandler) handleAddServer(cmd adapters.DiscordCommand) {
	if cmd.GuildID == "" {
		log.Printf("Add server command requires guild ID")
		return
	}

	guild := &entities.Guild{
		ID:       cmd.GuildID,
		Name:     "Unknown Guild", // We could fetch this from Discord API
		Channels: []string{},
	}

	if err := h.guildRepo.SaveGuild(guild); err != nil {
		log.Printf("Error adding server %s: %v", cmd.GuildID, err)
	} else {
		log.Printf("Successfully added server %s", cmd.GuildID)
	}
}

// handleAddChannel adds a new Discord channel to a server
func (h *InstanceHandler) handleAddChannel(cmd adapters.DiscordCommand) {
	if cmd.ChannelID == "" || cmd.GuildID == "" {
		log.Printf("Add channel command requires both channel ID and guild ID")
		return
	}

	// Check if guild exists
	guild, err := h.guildRepo.GetGuild(cmd.GuildID)
	if err != nil {
		log.Printf("Guild %s not found, creating it", cmd.GuildID)
		guild = &entities.Guild{
			ID:       cmd.GuildID,
			Name:     "Unknown Guild",
			Channels: []string{},
		}
		if err := h.guildRepo.SaveGuild(guild); err != nil {
			log.Printf("Error creating guild: %v", err)
			return
		}
	}

	// Create channel
	channel := &entities.Channel{
		ID:      cmd.ChannelID,
		GuildID: cmd.GuildID,
		Name:    "Unknown Channel", // We could fetch this from Discord API
		Games:   []string{},
	}

	if err := h.channelRepo.SaveChannel(channel); err != nil {
		log.Printf("Error adding channel %s: %v", cmd.ChannelID, err)
		return
	}

	// Update guild's channel list
	guild.Channels = append(guild.Channels, cmd.ChannelID)
	if err := h.guildRepo.UpdateGuild(guild); err != nil {
		log.Printf("Error updating guild channel list: %v", err)
	} else {
		log.Printf("Successfully added channel %s to guild %s", cmd.ChannelID, cmd.GuildID)
	}
}

// handleRemoveServer removes a Discord server/guild
func (h *InstanceHandler) handleRemoveServer(cmd adapters.DiscordCommand) {
	if cmd.GuildID == "" {
		log.Printf("Remove server command requires guild ID")
		return
	}

	// Get all channels in the guild
	channels, err := h.channelRepo.GetChannelsByGuild(cmd.GuildID)
	if err != nil {
		log.Printf("Error getting channels for guild %s: %v", cmd.GuildID, err)
		return
	}

	// Remove all channels in the guild
	for _, channel := range channels {
		if err := h.channelRepo.DeleteChannel(channel.ID); err != nil {
			log.Printf("Error removing channel %s: %v", channel.ID, err)
		}
	}

	// Remove the guild
	if err := h.guildRepo.DeleteGuild(cmd.GuildID); err != nil {
		log.Printf("Error removing server %s: %v", cmd.GuildID, err)
	} else {
		log.Printf("Successfully removed server %s and all its channels", cmd.GuildID)
	}
}

// handleRemoveChannel removes a Discord channel
func (h *InstanceHandler) handleRemoveChannel(cmd adapters.DiscordCommand) {
	if cmd.ChannelID == "" {
		log.Printf("Remove channel command requires channel ID")
		return
	}

	// Get the channel to find its guild
	channel, err := h.channelRepo.GetChannel(cmd.ChannelID)
	if err != nil {
		log.Printf("Channel %s not found", cmd.ChannelID)
		return
	}

	// Remove the channel
	if err := h.channelRepo.DeleteChannel(cmd.ChannelID); err != nil {
		log.Printf("Error removing channel %s: %v", cmd.ChannelID, err)
		return
	}

	// Update guild's channel list
	guild, err := h.guildRepo.GetGuild(channel.GuildID)
	if err == nil {
		var newChannels []string
		for _, chID := range guild.Channels {
			if chID != cmd.ChannelID {
				newChannels = append(newChannels, chID)
			}
		}
		guild.Channels = newChannels
		h.guildRepo.UpdateGuild(guild)
	}

	log.Printf("Successfully removed channel %s", cmd.ChannelID)
}

// handleListChannels lists all channels in a guild
func (h *InstanceHandler) handleListChannels(cmd adapters.DiscordCommand) {
	if cmd.GuildID == "" {
		log.Printf("List channels command requires guild ID")
		return
	}

	channels, err := h.channelRepo.GetChannelsByGuild(cmd.GuildID)
	if err != nil {
		log.Printf("Error getting channels for guild %s: %v", cmd.GuildID, err)
		return
	}

	log.Printf("Channels in guild %s:", cmd.GuildID)
	for _, channel := range channels {
		log.Printf("  - %s (%s) - Tracking %d games",
			channel.Name, channel.ID, len(channel.Games))
	}
}

// AddGameToChannel adds a game to a channel's tracking list
func (h *InstanceHandler) AddGameToChannel(channelID string, gameID string) error {
	channel, err := h.channelRepo.GetChannel(channelID)
	if err != nil {
		return err
	}

	// Check if game is already being tracked
	for _, game := range channel.Games {
		if game == gameID {
			return nil // Already tracking
		}
	}

	channel.Games = append(channel.Games, gameID)
	return h.channelRepo.UpdateChannel(channel)
}

// RemoveGameFromChannel removes a game from a channel's tracking list
func (h *InstanceHandler) RemoveGameFromChannel(channelID string, gameID string) error {
	channel, err := h.channelRepo.GetChannel(channelID)
	if err != nil {
		return err
	}

	var newGames []string
	for _, game := range channel.Games {
		if game != gameID {
			newGames = append(newGames, game)
		}
	}

	channel.Games = newGames
	return h.channelRepo.UpdateChannel(channel)
}

// GetChannelsForGame returns all channels tracking a specific game
func (h *InstanceHandler) GetChannelsForGame(gameID string) ([]*entities.Channel, error) {
	return h.channelRepo.GetChannelsByGame(gameID)
}

// GetChannelsForGuild returns all channels in a guild
func (h *InstanceHandler) GetChannelsForGuild(guildID string) ([]*entities.Channel, error) {
	return h.channelRepo.GetChannelsByGuild(guildID)
}
