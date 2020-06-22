package main

import (
	"context"
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
	"github.com/SonarBeserk/sophie-go/internal/db"
	"github.com/SonarBeserk/sophie-go/internal/embed"
	"github.com/SonarBeserk/sophie-go/internal/helpers"
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

type contextKey string

// Variables used for command line parameters
var (
	Token        string
	emotesFile   string
	databaseFile string

	database *db.Database

	emotes map[string]embed.Func = map[string]embed.Func{
		"smug": embed.CreateSmugEmbed,
	}
	emoteImages map[string][]string = map[string][]string{}

	databaseCtx contextKey = "db"
)

func init() {
	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.StringVar(&emotesFile, "emotes", "./emotes.toml", "Path to file containing emotes")
	flag.StringVar(&databaseFile, "db", "./data.db", "Path to database")
	flag.Parse()
}

func main() {
	err := loadEmoteMaps(emotesFile)
	if err != nil {
		fmt.Printf("error loading emotes file %s: %v", emotesFile, err)
		return
	}

	db, err := db.OpenOrConfigureDatabase(databaseFile)
	if err != nil {
		fmt.Printf("error loading database file %s: %v", databaseFile, err)
	}

	database = db
	defer database.Close()

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

func loadEmoteMaps(path string) error {
	data, err := ioutil.ReadFile(path)
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

	isPrivate, err := helpers.IsPrivateChat(s, m.ChannelID)
	if err != nil {
		fmt.Printf("Error occurred verifying channel type %s %v", m.ChannelID, err)
	}

	if isPrivate {
		fmt.Println("Ignoring private chat")
		return
	}

	userName, err := helpers.GetUserName(s, m.GuildID, s.State.User.ID)
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

	message := ""

	if len(msgParts) > 2 {
		userName := msgParts[2]
		usr, err := helpers.GetUserByName(s, m.GuildID, userName)
		if err != nil {
			fmt.Printf("Error occurred getting username %s %v", userName, err)
			return
		}

		receiverUsr = usr
	}

	if len(msgParts) > 3 {
		message = strings.Join(msgParts[3:], " ")
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

	c := context.Background()
	ctx := context.WithValue(c, databaseCtx, database)

	embed, err := embedFunc(ctx, senderUsr, receiverUsr, image, message)
	if err != nil {
		fmt.Printf("Error occurred creating embed %v", err)
		return
	}

	_, err = s.ChannelMessageSendEmbed(m.ChannelID, embed)
	if err != nil {
		fmt.Printf("Error occurred sending embed %v", err)
		return
	}
}
