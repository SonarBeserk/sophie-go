package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/bwmarrin/discordgo"
)

// Config represents the configuration for the bot
type Config struct {
	Emotes []Emote `toml:"emote"`
}

// Emote represents a emote that has an image
type Emote struct {
	Verb string
	URL  string
}

// Variables used for command line parameters
var (
	Token     string
	userNames map[string]string    = map[string]string{}
	emotes    map[string]EmbedFunc = map[string]EmbedFunc{
		"smug": createSmugEmbed,
	}
	emoteImages map[string][]string = map[string][]string{}
)

func init() {
	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.Parse()
}

func main() {
	err := loadEmoteMaps()
	if err != nil {
		fmt.Println("error loading emotes file,", err)
		return
	}

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Printf("error creating Discord session: %v", err)
		return
	}

	// Register the messageCreate func as a callback for MessageCreate events.
	dg.AddHandler(messageCreate)

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Printf("error opening connection: %v", err)
		return
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()
}

func loadEmoteMaps() error {
	data, err := ioutil.ReadFile("./emotes.toml")
	if err != nil {
		return err
	}

	var conf Config
	if _, err := toml.Decode(string(data), &conf); err != nil {
		return err
	}

	for _, emote := range conf.Emotes {
		emoteImages[emote.Verb] = append(emoteImages[emote.Verb], emote.URL)
	}

	return nil
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the authenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}

	isPrivate, err := isPrivateChat(s, m.ChannelID)
	if err != nil {
		fmt.Printf("Error occurred verifying channel type %s %v", m.ChannelID, err)
	}

	if isPrivate {
		fmt.Println("Ignoring private chat")
		return
	}

	userName, err := getUserName(s, m.GuildID, s.State.User.ID)
	if err != nil {
		fmt.Printf("Error occurred determining guild username %s %v", m.GuildID, err)
	}

	if !strings.HasPrefix(m.Content, userName) {
		return
	}

	msgParts := strings.Split(m.Content, " ")

	if len(msgParts) < 2 {
		return
	}

	senderUsr, err := s.GuildMember(m.GuildID, m.Author.ID)
	if err != nil {
		fmt.Printf("Error occurred getting username %s %v", m.Author.ID, err)
		return
	}

	var receiverUsr *discordgo.Member

	if len(msgParts) > 2 {
		userName := msgParts[2]
		usr, err := getUserByName(s, m.GuildID, userName)
		if err != nil {
			fmt.Printf("Error occurred getting username %s %v", userName, err)
			return
		}

		receiverUsr = usr
	}

	emote := strings.ToLower(msgParts[1])
	embedFunc := emotes[emote]

	if embedFunc == nil {
		return
	}

	// Add randomness
	rand.Seed(time.Now().UnixNano())

	r := rand.Intn(len(emoteImages[emote]))

	image := emoteImages["smug"][r]
	embed := embedFunc(senderUsr, receiverUsr, image)

	_, err = s.ChannelMessageSendEmbed(m.ChannelID, embed)
	if err != nil {
		fmt.Printf("Error occurred sending embed %v", err)
		return
	}
}

func isPrivateChat(s *discordgo.Session, channelID string) (bool, error) {
	channel, err := s.State.Channel(channelID)
	if err != nil {
		if channel, err = s.Channel(channelID); err != nil {
			fmt.Printf("Error occurred getting channel %s %v", channelID, err)
			return true, err
		}
	}

	return channel.Type != discordgo.ChannelTypeGuildText, nil
}

func getUserName(s *discordgo.Session, guildID string, userID string) (string, error) {
	key := guildID + "|" + userID
	name, ok := userNames[key]

	if !ok {
		usr, err := s.GuildMember(guildID, userID)
		if err != nil {
			fmt.Printf("Error occurred getting username %s %v", userID, err)
			return "", err
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

func getUserByName(s *discordgo.Session, GuildID string, userName string) (*discordgo.Member, error) {
	members, err := s.GuildMembers(GuildID, "", 1000)
	if err != nil {
		return nil, err
	}

	for _, member := range members {
		if member.Nick == userName || member.User.Username == userName {
			return member, nil
		}
	}

	return nil, nil
}
