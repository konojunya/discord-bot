package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"context"
	"github.com/bwmarrin/discordgo"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"

	"google.golang.org/api/option"
)

var client *auth.Client
var ctx context.Context

func onMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	println(m.Content)

	if m.Author.ID == s.State.User.ID {
		return
	}

	// Hello
	if m.Content == fmt.Sprintf("<@%s> Hello", s.State.User.ID) {
		reference := m.Reference()
		s.ChannelMessageSendReply(m.ChannelID, fmt.Sprintf("Hello! %s", m.Author.Username), reference)
	}

	// send direct message when !connect command
	if m.Content == "!connect" {
		c, err := s.UserChannelCreate(m.Author.ID)

		if err != nil {
			s.ChannelMessageSend(
				m.ChannelID,
				"Something went wrong while sending the DM!",
			)
			return
		}

		_, err = s.ChannelMessageSend(c.ID, "FirebaseAuthTesting へようこそ！\n電話番号を入力してください。")
		println(c.ID)

		if err != nil {
			s.ChannelMessageSend(
				m.ChannelID,
				"Failed to send you a DM.\nDid you disable DM in your privacy settings?",
			)
			return
		}
	}

	// in User DM
	// 上記 c.ID で取得したDMのIDを指定
	if m.ChannelID == "971257348905635920" {
		// 電話番号の本人確認をする必要がある
		user, err := client.GetUserByPhoneNumber(ctx, m.Content)
		if err != nil {
			s.ChannelMessageSend(
				m.ChannelID,
				"Failed to get user.\nPlease try again.",
			)
			return
		}

		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s さん、こんにちは！", user.UID))
	}
}

func main() {
	opt := option.WithCredentialsFile("sa.json")
	ctx = context.Background()
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		fmt.Errorf("error initializing app: %v", err)
	}

	c, err := app.Auth(ctx)
	if err != nil {
		panic(err)
	}

	client = c
	discord, err := discordgo.New("Bot " + os.Getenv("DISCORD_TOKEN"))

	if err != nil {
		panic(err)
	}

	println("Bot is now running.  Press CTRL-C to exit.")

	discord.AddHandler(onMessageCreate)

	err = discord.Open()
	stopBot := make(chan os.Signal, 1)
	signal.Notify(stopBot, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)

	<-stopBot

	err = discord.Close()

	return
}
