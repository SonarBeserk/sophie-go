package commands

import (
	"context"

	"github.com/bwmarrin/discordgo"
)

// Func provides a function used to implement a command
type Func func(ctx context.Context, s *discordgo.Session, msgParts []string, guildID string, authorID string, channelID string) error
