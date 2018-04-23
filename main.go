package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/bwmarrin/discordgo"
	"github.com/dags-/CaptainEggplant/quotes"
	"bufio"
)

// rub: 233182426711588864
// invite: https://discordapp.com/api/oauth2/authorize?client_id=437870761781231617&permissions=2112&scope=bot

var target *string
var q *quotes.Quotes

func main() {
	// discord bot api token
	token := flag.String("token", "", "Auth token")
	// tumblr api key
	key := flag.String("key", "", "Tumblr api key")
	// discord user id
	target = flag.String("target", "", "Egg-plantee's id")
	flag.Parse()

	// check all flags provided
	if *token == "" || *target == "" || *key == "" {
		fmt.Printf("flags error: token='%s', key='%s', target='%s'\n", *token, *key, *target)
		return
	}

	// init quotes & discord client
	q = quotes.New(*key)
	s, e := discordgo.New("Bot " + *token)
	if e != nil {
		fmt.Println("login err:", e)
		return
	}

	// add event handlers
	s.AddHandler(join)
	s.AddHandler(message)

	// open connection
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
	fmt.Println("joined guild:", g.Name)
}

func message(s *discordgo.Session, m *discordgo.MessageCreate) {
	// message author is our target
	if m.Author.ID == *target {
		// eggplant that mofo
		e := s.MessageReactionAdd(m.ChannelID, m.Message.ID, "🍆")
		if e != nil {
			fmt.Println("add reaction error:", e)
		}

		// automatically reply to user if haven't done so in the last 30 mins
		respond := q.ShouldRespond()
		if !respond {
			// otherwise if target has mentioned the bot, respond directly
			for _, u := range m.Mentions {
				if u.ID == s.State.User.ID {
					respond = true
					break
				}
			}
		}

		if respond {
			// poll the next quote
			txt := q.NextQuote()
			if txt != "" {
				// Oh (mention), with your face like (text)
				content := fmt.Sprint("Oh ", m.Author.Mention(), ", with your face like ", txt, " :eggplant:")
				_, e = s.ChannelMessageSend(m.ChannelID, content)
				if e != nil {
					fmt.Println(e)
				}
			}
		}
	}
}
