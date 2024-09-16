package discord

import (
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/ftrbnd/film-sync/internal/util"
)

func Session() *discordgo.Session {
	token := util.LoadEnvVar("DISCORD_TOKEN")

	discord, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatalf("Unable to start Discord session: %v", err)
	}

	return discord
}

func SendMessage(content string) {
	discord := Session()
	defer discord.Close()

	userID := util.LoadEnvVar("DISCORD_USER_ID")

	channel, err := discord.UserChannelCreate(userID)
	if err != nil {
		log.Fatalf("Failed to create DM channel: %v", err)
	}

	_, err = discord.ChannelMessageSend(channel.ID, content)
	if err != nil {
		log.Fatalf("Failed to send message to user: %v", err)
	}
}
