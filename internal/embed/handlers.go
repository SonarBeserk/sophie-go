package embed

import (
	"context"
	"errors"
	"fmt"

	"github.com/SonarBeserk/sophie-go/internal/db"
	"github.com/SonarBeserk/sophie-go/internal/emote"
	"github.com/bwmarrin/discordgo"
)

var (
	databaseCtx ContextKey = "db"
)

// ContextKey is used to store a value in context
type ContextKey string

// CreateEmbed creates an embed
func CreateEmbed(ctx context.Context, em emote.Emote, sender *discordgo.Member, receiver *discordgo.Member, image string, message string) (*discordgo.MessageEmbed, error) {
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
		sentCount, receivedCount, err := db.GetEmoteCountsForUser(em.Verb, sender.User.ID)
		if err != nil {
			return nil, err
		}

		err = db.SetEmoteSentUsage(em.Verb, sender.User.ID, sentCount+1)
		if err != nil {
			return nil, err
		}

		message = fmt.Sprintf(em.SenderMessage, senderName)
		stats = fmt.Sprintf(em.SenderDescription, senderName, sentCount, receivedCount)
	}

	if sender != nil && receiver != nil {
		sentCount, err := db.GetEmoteSentUsage(em.Verb, sender.User.ID)
		if err != nil {
			return nil, err
		}

		err = db.SetEmoteSentUsage(em.Verb, sender.User.ID, sentCount+1)
		if err != nil {
			return nil, err
		}

		receivedCount, err := db.GetEmoteReceivedUsage(em.Verb, receiver.User.ID)
		if err != nil {
			return nil, err
		}

		err = db.SetEmoteReceivedUsage(em.Verb, receiver.User.ID, receivedCount+1)
		if err != nil {
			return nil, err
		}

		message = fmt.Sprintf(em.ReceiverMessage, senderName, receiverName)
		stats = fmt.Sprintf(em.ReceiverDescription, sentCount, receivedCount)
	}

	embed := NewEmbed().
		SetDescription(message).
		SetImage(image).
		SetFooter(stats).
		SetColor(0x00ff00).MessageEmbed

	return embed, nil
}
