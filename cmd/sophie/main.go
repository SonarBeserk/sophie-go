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
	"github.com/SonarBeserk/sophie-go/internal/emote"
	"github.com/SonarBeserk/sophie-go/internal/helpers"
	"github.com/bwmarrin/discordgo"
)

// Config represents the configuration for the bot
type Config struct {
	Emotes []emote.Emote `toml:"emote"`
	Gifs   []emote.Gif   `toml:"gif"`
}

// Variables used for command line parameters
var (
	Token        string
	emotesFile   string
	databaseFile string

	database *db.Database

	emotes      map[string]emote.Emote = map[string]emote.Emote{}
	emoteImages map[string][]string    = map[string][]string{}

	databaseCtx embed.ContextKey = "db"
)

func init() {
	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.StringVar(&emotesFile, "emotes", "./emotes.toml", "Path to file containing emotes")
	flag.StringVar(&databaseFile, "db", "./data.db", "Path to database")
	flag.Parse()
}

func main() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Fprintf(os.Stderr, "Exception: %v\n", err)
		}
	}()

	err := loadEmoteMaps(emotesFile)
	if err != nil {
		fmt.Printf("Error loading emotes file %s: %v\n", emotesFile, err)
		return
	}

	db, err := db.OpenOrConfigureDatabase(databaseFile)
	if err != nil {
		fmt.Printf("Error loading database file %s: %v\n", databaseFile, err)
	}

	database = db
	defer database.Close()

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Printf("Error creating Discord session: %v\n", err)
		return
	}

	// Register the messageCreate func as a callback for MessageCreate events.
	dg.AddHandler(messageCreate)

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Printf("Error opening connection: %v\n", err)
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
		emotes[emote.Verb] = emote
	}

	for _, gif := range conf.Gifs {
		emoteImages[gif.Verb] = append(emoteImages[gif.Verb], gif.URL)
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
		fmt.Printf("Error occurred verifying channel type %s %v\n", m.ChannelID, err)
	}

	if isPrivate {
		fmt.Println("Ignoring private chat")
		return
	}

	userName, err := helpers.GetUserName(s, m.GuildID, s.State.User.ID)
	if err != nil {
		fmt.Printf("Error occurred determining guild username %s %v\n", m.GuildID, err)
	}

	if !strings.HasPrefix(strings.ToLower(m.Content), strings.ToLower(userName)) {
		return
	}

	msgParts := strings.Split(m.Content, " ")

	if len(msgParts) < 2 {
		return
	}

	senderUsr, err := s.GuildMember(m.GuildID, m.Author.ID)
	if err != nil {
		fmt.Printf("Error occurred getting username %s %v\n", m.Author.ID, err)
		return
	}

	var receiverUsr *discordgo.Member

	message := ""

	if len(msgParts) > 2 {
		userName := msgParts[2]
		usr, err := helpers.GetUserByName(s, m.GuildID, userName, true)
		if err != nil {
			fmt.Printf("Error occurred getting username %s %v\n", userName, err)
			return
		}

		receiverUsr = usr
	}

	if len(msgParts) > 3 {
		message = strings.Join(msgParts[3:], " ")
	}

	emote := strings.ToLower(msgParts[1])

	// Add randomness
	rand.Seed(time.Now().UnixNano())

	if len(emoteImages[emote]) == 0 {
		return
	}

	r := rand.Intn(len(emoteImages[emote]))

	image := emoteImages[emote][r]
	emoteEntry := emotes[emote]

	c := context.Background()
	ctx := context.WithValue(c, databaseCtx, *database)

	embed, err := embed.CreateEmoteEmbed(ctx, emoteEntry, senderUsr, receiverUsr, image, message)
	if err != nil {
		fmt.Printf("Error occurred creating embed: %v\n", err)
		return
	}

	_, err = s.ChannelMessageSendEmbed(m.ChannelID, embed)
	if err != nil {
		fmt.Printf("Error occurred sending embed: %v\n", err)
		return
	}
}
