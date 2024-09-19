package discord

import (
	"github.com/bwmarrin/discordgo"
	"github.com/ftrbnd/film-sync/internal/util"
)

func Session() *discordgo.Session {
	token := util.LoadEnvVar("DISCORD_TOKEN")

	discord, err := discordgo.New("Bot " + token)
	util.CheckError("Unable to start Discord session", err)

	return discord
}

func SendMessage(content string) {
	discord := Session()
	defer discord.Close()

	userID := util.LoadEnvVar("DISCORD_USER_ID")

	channel, err := discord.UserChannelCreate(userID)
	util.CheckError("Failed to create DM channel", err)

	_, err = discord.ChannelMessageSend(channel.ID, content)
	util.CheckError("Failed to send message to user", err)
}
