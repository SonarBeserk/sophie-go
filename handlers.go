package main

import "github.com/bwmarrin/discordgo"

// EmbedFunc represents a method used to create message embeds
type EmbedFunc func(m *discordgo.Member) *discordgo.MessageEmbed

func createSmugEmbed(m *discordgo.Member) *discordgo.MessageEmbed {
	name := m.User.Username

	if m.Nick != "" {
		name = m.Nick
	}

	embed := NewEmbed().
		SetDescription("**" + name + "**" + " is feeling " + "**" + emote + "**").
		SetImage("").
		SetFooter(name + " has been smug " + "20 times").
		SetColor(0x00ff00).MessageEmbed

	return embed
}
