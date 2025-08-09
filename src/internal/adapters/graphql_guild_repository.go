package adapters

import (
	"encoding/json"
	"fmt"

	"github.com/18xxnotifier/internal/domain/entities"
)

// GraphQLGuildRepository implements entities.GuildRepository using GraphQL
type GraphQLGuildRepository struct {
	graphqlAdapter GraphQLAdapter
}

// NewGraphQLGuildRepository creates a new GraphQL guild repository
func NewGraphQLGuildRepository(graphqlAdapter GraphQLAdapter) *GraphQLGuildRepository {
	return &GraphQLGuildRepository{
		graphqlAdapter: graphqlAdapter,
	}
}

// SaveGuild saves a new guild
func (r *GraphQLGuildRepository) SaveGuild(guild *entities.Guild) error {
	mutation := `
		mutation AddGuild($guild: AddGuildInput!) {
			addGuild(input: [$guild]) {
				guild {
					id
				}
			}
		}
	`

	input := map[string]interface{}{
		"id":       guild.ID,
		"name":     guild.Name,
		"channels": guild.Channels,
	}

	variables := map[string]interface{}{
		"guild": input,
	}

	_, err := r.graphqlAdapter.Mutate(mutation, variables)
	if err != nil {
		return fmt.Errorf("failed to save guild: %w", err)
	}

	return nil
}

// GetGuild retrieves a guild by ID
func (r *GraphQLGuildRepository) GetGuild(guildID string) (*entities.Guild, error) {
	query := `
		query GetGuild($guildID: String!) {
			queryGuild(filter: { id: { eq: $guildID } }) {
				id
				name
				channels
			}
		}
	`

	variables := map[string]interface{}{
		"guildID": guildID,
	}

	result, err := r.graphqlAdapter.Query(query, variables)
	if err != nil {
		return nil, fmt.Errorf("failed to query guild: %w", err)
	}

	var response struct {
		Data struct {
			QueryGuild []struct {
				ID       string   `json:"id"`
				Name     string   `json:"name"`
				Channels []string `json:"channels"`
			} `json:"queryGuild"`
		} `json:"data"`
	}

	if err := json.Unmarshal(result, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(response.Data.QueryGuild) == 0 {
		return nil, fmt.Errorf("guild not found: %s", guildID)
	}

	guildData := response.Data.QueryGuild[0]
	guild := &entities.Guild{
		ID:       guildData.ID,
		Name:     guildData.Name,
		Channels: guildData.Channels,
	}

	return guild, nil
}

// UpdateGuild updates an existing guild
func (r *GraphQLGuildRepository) UpdateGuild(guild *entities.Guild) error {
	mutation := `
		mutation UpdateGuild($guild: UpdateGuildInput!) {
			updateGuild(input: $guild) {
				guild {
					id
				}
			}
		}
	`

	input := map[string]interface{}{
		"filter": map[string]interface{}{
			"id": map[string]interface{}{
				"eq": guild.ID,
			},
		},
		"set": map[string]interface{}{
			"name":     guild.Name,
			"channels": guild.Channels,
		},
	}

	variables := map[string]interface{}{
		"guild": input,
	}

	_, err := r.graphqlAdapter.Mutate(mutation, variables)
	if err != nil {
		return fmt.Errorf("failed to update guild: %w", err)
	}

	return nil
}

// DeleteGuild deletes a guild by ID
func (r *GraphQLGuildRepository) DeleteGuild(guildID string) error {
	mutation := `
		mutation DeleteGuild($filter: GuildFilter!) {
			deleteGuild(filter: $filter) {
				msg
			}
		}
	`

	filter := map[string]interface{}{
		"id": map[string]interface{}{
			"eq": guildID,
		},
	}

	variables := map[string]interface{}{
		"filter": filter,
	}

	_, err := r.graphqlAdapter.Mutate(mutation, variables)
	if err != nil {
		return fmt.Errorf("failed to delete guild: %w", err)
	}

	return nil
}

// GetAllGuilds retrieves all guilds
func (r *GraphQLGuildRepository) GetAllGuilds() ([]*entities.Guild, error) {
	query := `
		query GetAllGuilds {
			queryGuild {
				id
				name
				channels
			}
		}
	`

	result, err := r.graphqlAdapter.Query(query, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to query guilds: %w", err)
	}

	var response struct {
		Data struct {
			QueryGuild []struct {
				ID       string   `json:"id"`
				Name     string   `json:"name"`
				Channels []string `json:"channels"`
			} `json:"queryGuild"`
		} `json:"data"`
	}

	if err := json.Unmarshal(result, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	var guilds []*entities.Guild
	for _, guildData := range response.Data.QueryGuild {
		guild := &entities.Guild{
			ID:       guildData.ID,
			Name:     guildData.Name,
			Channels: guildData.Channels,
		}
		guilds = append(guilds, guild)
	}

	return guilds, nil
}
