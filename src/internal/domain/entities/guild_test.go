package entities

import (
	"testing"
)

func TestGuild_HasChannel(t *testing.T) {
	tests := []struct {
		name      string
		guild     *Guild
		channelID string
		expected  bool
	}{
		{
			name: "guild has channel",
			guild: &Guild{
				ID:       "guild1",
				Name:     "Test Guild",
				Channels: []string{"channel1", "channel2"},
			},
			channelID: "channel1",
			expected:  true,
		},
		{
			name: "guild does not have channel",
			guild: &Guild{
				ID:       "guild1",
				Name:     "Test Guild",
				Channels: []string{"channel1", "channel2"},
			},
			channelID: "channel3",
			expected:  false,
		},
		{
			name: "guild has no channels",
			guild: &Guild{
				ID:       "guild1",
				Name:     "Test Guild",
				Channels: []string{},
			},
			channelID: "channel1",
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.guild.HasChannel(tt.channelID)
			if result != tt.expected {
				t.Errorf("Guild.HasChannel(%s) = %v, want %v", tt.channelID, result, tt.expected)
			}
		})
	}
}

func TestGuild_AddChannel(t *testing.T) {
	guild := &Guild{
		ID:       "guild1",
		Name:     "Test Guild",
		Channels: []string{"channel1"},
	}

	// Test adding new channel
	guild.AddChannel("channel2")
	expected := []string{"channel1", "channel2"}

	if len(guild.Channels) != len(expected) {
		t.Errorf("Guild.AddChannel() length = %d, want %d", len(guild.Channels), len(expected))
	}

	for i, channel := range guild.Channels {
		if channel != expected[i] {
			t.Errorf("Guild.AddChannel()[%d] = %s, want %s", i, channel, expected[i])
		}
	}

	// Test adding duplicate channel (should not add)
	guild.AddChannel("channel1")
	if len(guild.Channels) != len(expected) {
		t.Errorf("Guild.AddChannel() duplicate length = %d, want %d", len(guild.Channels), len(expected))
	}
}

func TestGuild_RemoveChannel(t *testing.T) {
	guild := &Guild{
		ID:       "guild1",
		Name:     "Test Guild",
		Channels: []string{"channel1", "channel2", "channel3"},
	}

	// Test removing existing channel
	guild.RemoveChannel("channel2")
	expected := []string{"channel1", "channel3"}

	if len(guild.Channels) != len(expected) {
		t.Errorf("Guild.RemoveChannel() length = %d, want %d", len(guild.Channels), len(expected))
	}

	for i, channel := range guild.Channels {
		if channel != expected[i] {
			t.Errorf("Guild.RemoveChannel()[%d] = %s, want %s", i, channel, expected[i])
		}
	}

	// Test removing non-existent channel
	guild.RemoveChannel("channel4")
	if len(guild.Channels) != len(expected) {
		t.Errorf("Guild.RemoveChannel() non-existent length = %d, want %d", len(guild.Channels), len(expected))
	}
}

func TestGuild_GetChannels(t *testing.T) {
	guild := &Guild{
		ID:       "guild1",
		Name:     "Test Guild",
		Channels: []string{"channel1", "channel2", "channel3"},
	}

	result := guild.GetChannels()
	expected := []string{"channel1", "channel2", "channel3"}

	if len(result) != len(expected) {
		t.Errorf("Guild.GetChannels() length = %d, want %d", len(result), len(expected))
	}

	for i, channel := range result {
		if channel != expected[i] {
			t.Errorf("Guild.GetChannels()[%d] = %s, want %s", i, channel, expected[i])
		}
	}
}

func TestGuild_GetName(t *testing.T) {
	guild := &Guild{
		ID:       "guild1",
		Name:     "Test Guild",
		Channels: []string{"channel1"},
	}

	result := guild.GetName()
	expected := "Test Guild"

	if result != expected {
		t.Errorf("Guild.GetName() = %s, want %s", result, expected)
	}
}

func TestGuild_GetID(t *testing.T) {
	guild := &Guild{
		ID:       "guild1",
		Name:     "Test Guild",
		Channels: []string{"channel1"},
	}

	result := guild.GetID()
	expected := "guild1"

	if result != expected {
		t.Errorf("Guild.GetID() = %s, want %s", result, expected)
	}
}
