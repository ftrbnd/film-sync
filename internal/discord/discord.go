package discord

import (
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/ftrbnd/film-sync/internal/util"
)

func Session() *discordgo.Session {
	token := util.LoadEnvVar("DISCORD_TOKEN")

	s, err := discordgo.New("Bot " + token)
	util.CheckError("Unable to start Discord session", err)

	err = s.Open()
	util.CheckError("Failed to open Discord session", err)

	s.AddHandler(func(s *discordgo.Session, m *discordgo.InteractionCreate) {
		log.Default().Println("EVENT:", m.Type)
	})

	log.Default().Printf("[Discord] %s is ready", s.State.User)
	return s
}

func createDMChannel(s *discordgo.Session) *discordgo.Channel {
	userID := util.LoadEnvVar("DISCORD_USER_ID")

	c, err := s.UserChannelCreate(userID)
	util.CheckError("Failed to create DM channel", err)

	return c
}

func SendAuthMessage(authURL string, s *discordgo.Session) {
	channel := createDMChannel(s)

	_, err := s.ChannelMessageSendComplex(channel.ID, &discordgo.MessageSend{
		Embeds: []*discordgo.MessageEmbed{
			{
				Title:       "Authentication required!",
				Description: "Visit the link to connect with Gmail and Google Drive",
				Color:       0xFFFB25,
				URL:         "https://fly.io/apps/film-sync/monitoring",
			},
		},
		Components: []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.Button{
						Label: "Sign in with Google",
						Style: discordgo.LinkButton,
						URL:   authURL,
					},
				},
			},
		},
	})
	util.CheckError("Failed to send auth message", err)
}

func SendSuccessMessage(s3Url string, driveUrl string, message string, s *discordgo.Session) {
	channel := createDMChannel(s)

	_, err := s.ChannelMessageSendComplex(channel.ID, &discordgo.MessageSend{
		Embeds: []*discordgo.MessageEmbed{
			{
				Title:       "Upload successful!",
				Description: message,
				Color:       0x32FF25,
				URL:         "https://fly.io/apps/film-sync/monitoring",
			},
		},
		Components: []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.Button{
						Label: "Google Drive",
						Style: discordgo.LinkButton,
						URL:   driveUrl,
					},
					discordgo.Button{
						Label: "AWS S3",
						Style: discordgo.LinkButton,
						URL:   s3Url,
					},
				},
			},
		},
	})
	util.CheckError("Failed to send success message", err)
}
