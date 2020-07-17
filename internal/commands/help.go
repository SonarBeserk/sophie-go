package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

// HandleListEmotes handles running commands
func HandleListEmotes(ctx context.Context, s *discordgo.Session, msgParts []string, guildID string, authorID string, channelID string) error {
	keys := make([]string, 0, len(emotes))

	for emote := range emotes {
		keys = append(keys, emote)
	}

	_, err := s.ChannelMessageSend(channelID, "Available Emotes: "+strings.Join(keys, ", "))
	if err != nil {
		return fmt.Errorf("error occurred sending embed: %v", err)
	}

	return nil
}
