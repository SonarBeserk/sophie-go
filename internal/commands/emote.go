package commands

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/SonarBeserk/sophie-go/internal/embed"
	"github.com/SonarBeserk/sophie-go/internal/emote"
	"github.com/SonarBeserk/sophie-go/internal/helpers"
	"github.com/bwmarrin/discordgo"
)

var (
	emotes      map[string]emote.Emote = map[string]emote.Emote{}
	emoteImages map[string][]string    = map[string][]string{}
)

// HandleEmote handles running commands
func HandleEmote(ctx context.Context, s *discordgo.Session, msgParts []string, guildID string, authorID string, channelID string) error {
	if len(msgParts) < 1 {
		return nil
	}

	emote := msgParts[0]

	senderUsr, err := s.GuildMember(guildID, authorID)
	if err != nil {
		return fmt.Errorf("error occurred getting username %s %v", authorID, err)
	}

	var receiverUsr *discordgo.Member

	message := ""

	if len(msgParts) > 1 {
		userName := msgParts[1]
		usr, err := helpers.GetUserByName(s, guildID, userName, true)
		if err != nil {
			return fmt.Errorf("error occurred getting user by name %s %v", userName, err)
		}

		receiverUsr = usr
	}

	if len(msgParts) > 2 {
		message = strings.Join(msgParts[2:], " ")
	}

	// Add randomness
	rand.Seed(time.Now().UnixNano())

	if len(emoteImages[emote]) == 0 {
		return nil
	}

	r := rand.Intn(len(emoteImages[emote]))

	image := emoteImages[emote][r]
	emoteEntry := emotes[emote]

	embed, err := embed.CreateEmoteEmbed(ctx, emoteEntry, senderUsr, receiverUsr, image, message)
	if err != nil {
		return fmt.Errorf("error occurred creating embed: %v", err)
	}

	_, err = s.ChannelMessageSendEmbed(channelID, embed)
	if err != nil {
		return fmt.Errorf("error occurred sending embed: %v", err)
	}

	return nil
}

// AddEmote adds an entry to the emotes list
func AddEmote(emote emote.Emote) {
	emotes[emote.Verb] = emote
}

// AddEmoteImage adds an image for an emote
func AddEmoteImage(gif emote.Gif) {
	emoteImages[gif.Verb] = append(emoteImages[gif.Verb], gif.URL)
}
