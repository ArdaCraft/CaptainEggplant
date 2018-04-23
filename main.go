package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/bwmarrin/discordgo"
)

// rub 233182426711588864
// https://discordapp.com/oauth2/authorize?client_id=437870761781231617&permissions=64&scope=bot

var token, target *string

func main() {
	token = flag.String("token", "", "Auth token")
	target = flag.String("target", "", "The egg plantee")
	if *token == "" || *target == "" {
		fmt.Println("token or id not provided: token=", "'" + *token + "'", "target=", "'" + *target + "'")
		return
	}

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

	c := make(chan os.Signal, 1)
	<-c
	s.Close()
}

func join(s *discordgo.Session, g *discordgo.GuildCreate) {
	fmt.Println("joined guild:", g.Name)
}

func message(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == *target {
		e := s.MessageReactionAdd(m.ChannelID, m.Message.ID, "🍆")
		if e != nil {
			fmt.Println("add reaction error:", e)
		}
	}
}
