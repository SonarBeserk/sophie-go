package main

import (
	"github.com/bwmarrin/discordgo"
)

// EmbedFunc represents a method used to create message embeds
type EmbedFunc func(sender *discordgo.Member, receiver *discordgo.Member, image string, message string) (*discordgo.MessageEmbed, error)

var (
	smugKey = "smug"
)

func createSmugEmbed(sender *discordgo.Member, receiver *discordgo.Member, image string, message string) (*discordgo.MessageEmbed, error) {
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
		sentCount, receivedCount, err := getEmoteCountsForUser(smugKey, sender.User.ID)
		if err != nil {
			return nil, err
		}

		message = "**" + senderName + "**" + " is feeling " + "**Smug**"
		stats = senderName + " has been smug " + sentCount + " times and has been treated smugly " + receivedCount + " times"
	}

	if sender != nil && receiver != nil {
		sentCount, receivedCount, err := getEmoteCountsForUser(smugKey, receiver.User.ID)
		if err != nil {
			return nil, err
		}

		message = "**" + senderName + "**" + " is feeling " + "**Smug** towards **" + receiverName + "**"
		stats = receiverName + " has been smug " + sentCount + " times and has been treated smugly " + receivedCount + " times"
	}

	embed := NewEmbed().
		SetDescription(message).
		SetImage(image).
		SetFooter(stats).
		SetColor(0x00ff00).MessageEmbed

	return embed, nil
}
