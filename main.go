package main

import (
	"bufio"
	"flag"
	"fmt"
	"math/rand"
	"os"

	"github.com/ArdaCraft/CaptainEggplant/quote"
	"github.com/bwmarrin/discordgo"
	"time"
	"github.com/dags-/discordapp/bot"
	"sync"
	"github.com/ArdaCraft/CaptainEggplant/plant"
	"github.com/dags-/discordapp/command"
	"github.com/dags-/discordapp/util"
)

// rub: 233182426711588864
// invite: https://discordapp.com/api/oauth2/authorize?client_id=437870761781231617&permissions=2112&scope=bot

var lock sync.RWMutex
var quotes *quote.Quotes
var plants *plant.Plants

func init()  {
	quotes = quote.New()
	plants = plant.New()
	plants.Save()
}

func main() {
	token := flag.String("token", "", "Auth token")
	flag.Parse()

	// check all flags provided
	if *token == "" {
		fmt.Printf("no token provided")
		return
	}

	b := bot.New(token)
	b.AddHandler(message)
	b.AddCommand(command.New("!egg set <@user>", &[]string{"Admin", "Developer"}, setPlant))
	b.AddCommand(command.New("!egg add <@user>", &[]string{"Admin", "Developer"}, addPlant))
	b.AddCommand(command.New("!egg rem <@user>", &[]string{"Admin", "Developer"}, remPlant))

	// listen for console 'stop' command
	go handleStop()

	// listen for kill signal then close the connection
	b.Connect()
}

func setPlant(ctx *command.Context) error {
	lock.Lock()
	defer lock.Unlock()
	user := ctx.Args["user"]
	if user != "" {
		plants.Main = user
		plants.All[user] = true
		plants.Save()
		fmt.Println("set main plant", user)
		ctx.Session.ChannelMessageDelete(ctx.Message.ChannelID, ctx.Message.ID)
	}
	return nil
}

func addPlant(ctx *command.Context) error {
	lock.Lock()
	defer lock.Unlock()
	user := ctx.Args["user"]
	if user != "" {
		plants.All[user] = true
		plants.Save()
		fmt.Println("added plant", user)
		ctx.Session.ChannelMessageDelete(ctx.Message.ChannelID, ctx.Message.ID)
	}
	return nil
}

func remPlant(ctx *command.Context) error {
	lock.Lock()
	defer lock.Unlock()
	user := ctx.Args["user"]
	if user != "" {
		if user == plants.Main {
			plants.Main = ""
		}
		delete(plants.All, user)
		fmt.Println("removed plant", user)
		ctx.Session.ChannelMessageDelete(ctx.Message.ChannelID, ctx.Message.ID)
	}
	return nil
}

func message(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.Bot || m.Author.ID == s.State.User.ID {
		return
	}

	lock.Lock()
	defer lock.Unlock()

	if _, ok := plants.All[m.Author.ID]; ok {
		e := s.MessageReactionAdd(m.ChannelID, m.Message.ID, "🍆")
		if e != nil {
			fmt.Println("add reaction error:", e)
		}
	}

	// @mentioning the bot invokes a response otherwise randomly send a message (less frequently)
	if util.Mentions(m, s.State.User.ID) {
		// rate limited to once every 15 secs
		if !quotes.CanInvoke(15 * time.Second) {
			return
		}

		sendResponse(s, m, quotes.NextResponse())
	} else {
		// rate limited to once every 12 hours
		if !quotes.CanRespond(6 * time.Hour) {
			return
		}

		// 5% chance of sending a message
		if rand.Intn(100) < 5 {
			sendResponse(s, m, quotes.NextResponse())
		} else {
			// set timestamp and try again in 12 hours
			quotes.Cooldown()
		}
	}
}

func sendResponse(s *discordgo.Session, m *discordgo.MessageCreate, msg string) {
	if msg == "" {
		fmt.Println("send empty message?")
		return
	}

	u, e := s.User(plants.Main)
	if e != nil {
		fmt.Println("get user error:", e)
		return
	}

	content := fmt.Sprint("Oh ", u.Mention(), ", with your face like a ", msg, " :eggplant:")
	_, e = s.ChannelMessageSend(m.ChannelID, content)

	if e != nil {
		fmt.Println("send response error:", e)
	}
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