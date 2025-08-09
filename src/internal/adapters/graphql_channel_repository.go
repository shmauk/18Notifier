package adapters

import (
	"encoding/json"
	"fmt"

	"github.com/18xxnotifier/internal/domain/entities"
)

// GraphQLChannelRepository implements entities.ChannelRepository using GraphQL
type GraphQLChannelRepository struct {
	graphqlAdapter GraphQLAdapter
}

// NewGraphQLChannelRepository creates a new GraphQL channel repository
func NewGraphQLChannelRepository(graphqlAdapter GraphQLAdapter) *GraphQLChannelRepository {
	return &GraphQLChannelRepository{
		graphqlAdapter: graphqlAdapter,
	}
}

// SaveChannel saves a new channel
func (r *GraphQLChannelRepository) SaveChannel(channel *entities.Channel) error {
	mutation := `
		mutation AddChannel($channel: AddChannelInput!) {
			addChannel(input: [$channel]) {
				channel {
					id
				}
			}
		}
	`

	input := map[string]interface{}{
		"id":      channel.ID,
		"guildId": channel.GuildID,
		"name":    channel.Name,
		"games":   channel.Games,
	}

	variables := map[string]interface{}{
		"channel": input,
	}

	_, err := r.graphqlAdapter.Mutate(mutation, variables)
	if err != nil {
		return fmt.Errorf("failed to save channel: %w", err)
	}

	return nil
}

// GetChannel retrieves a channel by ID
func (r *GraphQLChannelRepository) GetChannel(channelID string) (*entities.Channel, error) {
	query := `
		query GetChannel($channelID: String!) {
			queryChannel(filter: { id: { eq: $channelID } }) {
				id
				guildId
				name
				games
			}
		}
	`

	variables := map[string]interface{}{
		"channelID": channelID,
	}

	result, err := r.graphqlAdapter.Query(query, variables)
	if err != nil {
		return nil, fmt.Errorf("failed to query channel: %w", err)
	}

	var response struct {
		Data struct {
			QueryChannel []struct {
				ID      string   `json:"id"`
				GuildID string   `json:"guildId"`
				Name    string   `json:"name"`
				Games   []string `json:"games"`
			} `json:"queryChannel"`
		} `json:"data"`
	}

	if err := json.Unmarshal(result, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(response.Data.QueryChannel) == 0 {
		return nil, fmt.Errorf("channel not found: %s", channelID)
	}

	channelData := response.Data.QueryChannel[0]
	channel := &entities.Channel{
		ID:      channelData.ID,
		GuildID: channelData.GuildID,
		Name:    channelData.Name,
		Games:   channelData.Games,
	}

	return channel, nil
}

// UpdateChannel updates an existing channel
func (r *GraphQLChannelRepository) UpdateChannel(channel *entities.Channel) error {
	mutation := `
		mutation UpdateChannel($channel: UpdateChannelInput!) {
			updateChannel(input: $channel) {
				channel {
					id
				}
			}
		}
	`

	input := map[string]interface{}{
		"filter": map[string]interface{}{
			"id": map[string]interface{}{
				"eq": channel.ID,
			},
		},
		"set": map[string]interface{}{
			"guildId": channel.GuildID,
			"name":    channel.Name,
			"games":   channel.Games,
		},
	}

	variables := map[string]interface{}{
		"channel": input,
	}

	_, err := r.graphqlAdapter.Mutate(mutation, variables)
	if err != nil {
		return fmt.Errorf("failed to update channel: %w", err)
	}

	return nil
}

// DeleteChannel deletes a channel by ID
func (r *GraphQLChannelRepository) DeleteChannel(channelID string) error {
	mutation := `
		mutation DeleteChannel($filter: ChannelFilter!) {
			deleteChannel(filter: $filter) {
				msg
			}
		}
	`

	filter := map[string]interface{}{
		"id": map[string]interface{}{
			"eq": channelID,
		},
	}

	variables := map[string]interface{}{
		"filter": filter,
	}

	_, err := r.graphqlAdapter.Mutate(mutation, variables)
	if err != nil {
		return fmt.Errorf("failed to delete channel: %w", err)
	}

	return nil
}

// GetChannelsByGuild gets all channels in a guild
func (r *GraphQLChannelRepository) GetChannelsByGuild(guildID string) ([]*entities.Channel, error) {
	query := `
		query GetChannelsByGuild($guildID: String!) {
			queryChannel(filter: { guildId: { eq: $guildID } }) {
				id
				guildId
				name
				games
			}
		}
	`

	variables := map[string]interface{}{
		"guildID": guildID,
	}

	result, err := r.graphqlAdapter.Query(query, variables)
	if err != nil {
		return nil, fmt.Errorf("failed to query channels by guild: %w", err)
	}

	var response struct {
		Data struct {
			QueryChannel []struct {
				ID      string   `json:"id"`
				GuildID string   `json:"guildId"`
				Name    string   `json:"name"`
				Games   []string `json:"games"`
			} `json:"queryChannel"`
		} `json:"data"`
	}

	if err := json.Unmarshal(result, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	var channels []*entities.Channel
	for _, channelData := range response.Data.QueryChannel {
		channel := &entities.Channel{
			ID:      channelData.ID,
			GuildID: channelData.GuildID,
			Name:    channelData.Name,
			Games:   channelData.Games,
		}
		channels = append(channels, channel)
	}

	return channels, nil
}

// GetChannelsByGame gets all channels tracking a specific game
func (r *GraphQLChannelRepository) GetChannelsByGame(gameID string) ([]*entities.Channel, error) {
	query := `
		query GetChannelsByGame($gameID: String!) {
			queryChannel(filter: { games: { anyofterms: $gameID } }) {
				id
				guildId
				name
				games
			}
		}
	`

	variables := map[string]interface{}{
		"gameID": gameID,
	}

	result, err := r.graphqlAdapter.Query(query, variables)
	if err != nil {
		return nil, fmt.Errorf("failed to query channels by game: %w", err)
	}

	var response struct {
		Data struct {
			QueryChannel []struct {
				ID      string   `json:"id"`
				GuildID string   `json:"guildId"`
				Name    string   `json:"name"`
				Games   []string `json:"games"`
			} `json:"queryChannel"`
		} `json:"data"`
	}

	if err := json.Unmarshal(result, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	var channels []*entities.Channel
	for _, channelData := range response.Data.QueryChannel {
		channel := &entities.Channel{
			ID:      channelData.ID,
			GuildID: channelData.GuildID,
			Name:    channelData.Name,
			Games:   channelData.Games,
		}
		channels = append(channels, channel)
	}

	return channels, nil
}

// GetAllChannels retrieves all channels
func (r *GraphQLChannelRepository) GetAllChannels() ([]*entities.Channel, error) {
	query := `
		query GetAllChannels {
			queryChannel {
				id
				guildId
				name
				games
			}
		}
	`

	result, err := r.graphqlAdapter.Query(query, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to query all channels: %w", err)
	}

	var response struct {
		Data struct {
			QueryChannel []struct {
				ID      string   `json:"id"`
				GuildID string   `json:"guildId"`
				Name    string   `json:"name"`
				Games   []string `json:"games"`
			} `json:"queryChannel"`
		} `json:"data"`
	}

	if err := json.Unmarshal(result, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	var channels []*entities.Channel
	for _, channelData := range response.Data.QueryChannel {
		channel := &entities.Channel{
			ID:      channelData.ID,
			GuildID: channelData.GuildID,
			Name:    channelData.Name,
			Games:   channelData.Games,
		}
		channels = append(channels, channel)
	}

	return channels, nil
}
