package helpers

import (
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

var (
	userNames map[string]string = map[string]string{}
)

func IsPrivateChat(s *discordgo.Session, channelID string) (bool, error) {
	channel, err := s.State.Channel(channelID)
	if err != nil {
		if channel, err = s.Channel(channelID); err != nil {
			return true, errors.Wrapf(err, "Error occurred getting channel %s %v", channelID, err)
		}
	}

	return channel.Type != discordgo.ChannelTypeGuildText, nil
}

func GetUserName(s *discordgo.Session, guildID string, userID string) (string, error) {
	key := guildID + "|" + userID
	name, ok := userNames[key]

	if !ok {
		usr, err := s.GuildMember(guildID, userID)
		if err != nil {
			return "", errors.Wrapf(err, "Error occurred getting username %s %v", userID, err)
		}

		if usr.Nick != "" {
			userNames[key] = usr.Nick
		} else {
			userNames[key] = usr.User.Username
		}

		name = userNames[key]
	}

	return name, nil
}

func GetUserByName(s *discordgo.Session, GuildID string, userName string, fuzzy bool) (*discordgo.Member, error) {
	members, err := s.GuildMembers(GuildID, "", 1000)
	if err != nil {
		return nil, err
	}

	for _, member := range members {
		if fuzzy && strings.HasPrefix(member.Nick, userName) || fuzzy && strings.HasPrefix(member.User.Username, userName) {
			return member, nil
		}

		if member.Nick == userName || member.User.Username == userName {
			return member, nil
		}
	}

	return nil, nil
}
