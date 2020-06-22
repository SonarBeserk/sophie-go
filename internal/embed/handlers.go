package embed

import (
	"context"
	"errors"
	"strconv"

	"github.com/SonarBeserk/sophie-go/internal/db"
	"github.com/bwmarrin/discordgo"
)

type contextKey string

var (
	databaseCtx contextKey = "db"
)

// Func represents a method used to create message embeds
type Func func(ctx context.Context, sender *discordgo.Member, receiver *discordgo.Member, image string, message string) (*discordgo.MessageEmbed, error)

var (
	smugKey = "smug"
)

func CreateSmugEmbed(ctx context.Context, sender *discordgo.Member, receiver *discordgo.Member, image string, message string) (*discordgo.MessageEmbed, error) {
	db, ok := ctx.Value(databaseCtx).(db.Database)
	if !ok {
		return nil, errors.New("Failed to get database from context")
	}

	senderName := sender.User.Username

	if sender.Nick != "" {
		senderName = sender.Nick
	}

	receiverName := ""
	if receiver != nil {
		if receiver.Nick != "" {
			receiverName = receiver.Nick
		} else {
			receiverName = receiver.User.Username
		}
	}

	stats := ""

	if sender != nil && receiver == nil {
		sentCount, receivedCount, err := db.GetEmoteCountsForUser(smugKey, sender.User.ID)
		if err != nil {
			return nil, err
		}

		err = db.SetEmoteSentUsage(smugKey, sender.User.ID, sentCount+1)
		if err != nil {
			return nil, err
		}

		message = "**" + senderName + "**" + " is feeling " + "**Smug**"
		stats = senderName + " has been smug " + strconv.Itoa(sentCount) + " times and has been treated smugly " + strconv.Itoa(receivedCount) + " times"
	}

	if sender != nil && receiver != nil {
		sentCount, err := db.GetEmoteSentUsage(smugKey, sender.User.ID)
		if err != nil {
			return nil, err
		}

		err = db.SetEmoteSentUsage(smugKey, sender.User.ID, sentCount+1)
		if err != nil {
			return nil, err
		}

		receivedCount, err := db.GetEmoteReceivedUsage(smugKey, receiver.User.ID)
		if err != nil {
			return nil, err
		}

		err = db.SetEmoteReceivedUsage(smugKey, receiver.User.ID, receivedCount+1)
		if err != nil {
			return nil, err
		}

		message = "**" + senderName + "**" + " is feeling " + "**Smug** towards **" + receiverName + "**"
		stats = receiverName + " has been smug " + strconv.Itoa(sentCount) + " times and has been treated smugly " + strconv.Itoa(receivedCount) + " times"
	}

	embed := NewEmbed().
		SetDescription(message).
		SetImage(image).
		SetFooter(stats).
		SetColor(0x00ff00).MessageEmbed

	return embed, nil
}
