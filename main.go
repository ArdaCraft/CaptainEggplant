package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/bwmarrin/discordgo"
	"bufio"
	"github.com/dags-/CaptainEggplant/quotes"
	"math/rand"
)

// rub: 233182426711588864
// invite: https://discordapp.com/api/oauth2/authorize?client_id=437870761781231617&permissions=2112&scope=bot

var target *string
var q = quotes.New()

func main() {
	token := flag.String("token", "", "Auth token")
	target = flag.String("target", "", "Egg-plantee's id")
	flag.Parse()

	// check all flags provided
	if *token == "" || *target == "" {
		fmt.Printf("flags error: token='%s', target='%s'\n", *token, *target)
		return
	}

	// init discord client
	s, e := discordgo.New("Bot " + *token)
	if e != nil {
		fmt.Println("login err:", e)
		return
	}

	s.AddHandler(join)
	s.AddHandler(message)

	e = s.Open()
	if e != nil {
		fmt.Println("could not open session:", e)
		return
	}

	// listen for console 'stop' command
	go handleStop()

	// listen for kill signal then close the connection
	c := make(chan os.Signal, 1)
	<-c
	s.Close()
}

func handleStop() {
	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		if scanner.Text() == "stop" {
			fmt.Println("Stopping...")
			os.Exit(0)
			break
		}
	}
}

func join(s *discordgo.Session, g *discordgo.GuildCreate) {
	s.GuildMemberNickname(g.ID, s.State.User.ID, "Captain Eggplant 🍆")
	fmt.Println("joined guild:", g.Name)
}

func message(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.Bot || m.Author.ID == s.State.User.ID {
		return
	}

	// add reaction to target
	if m.Author.ID == *target {
		e := s.MessageReactionAdd(m.ChannelID, m.Message.ID, "🍆")
		if e != nil {
			fmt.Println("add reaction error:", e)
		}
	}

	// @mentioning the bot invokes a response otherwise randomly send a message (less frequently)
	if mentions(s.State.User.ID, m) {
		// rate limited to once every 15 secs
		if !q.CanInvoke() {
			return
		}

		sendResponse(s, m, q.NextResponse())
	} else {
		// rate limited to once every 12 hours
		if !q.CanRespond() {
			return
		}

		// 5% chance of sending a message
		if rand.Intn(100) < 5 {
			sendResponse(s, m, q.NextResponse())
		} else {
			// set timestamp and try again in 12 hours
			q.Cooldown()
		}
	}
}

func sendResponse(s *discordgo.Session, m *discordgo.MessageCreate, msg string) {
	if msg == "" {
		fmt.Println("send empty message?")
		return
	}

	content := fmt.Sprint("Oh ", m.Author.Mention(), ", with your face like a ", msg, " :eggplant:")
	_, e := s.ChannelMessageSend(m.ChannelID, content)

	if e != nil {
		fmt.Println("send response error:", e)
	}
}

func mentions(id string, m *discordgo.MessageCreate) bool {
	for _, a := range m.Mentions {
		if a.ID == id {
			return true
		}
	}
	return false
}