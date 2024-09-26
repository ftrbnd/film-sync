package discord

import (
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/ftrbnd/film-sync/internal/util"
)

func Session() *discordgo.Session {
	token := util.LoadEnvVar("DISCORD_TOKEN")

	s, err := discordgo.New("Bot " + token)
	util.CheckError("Unable to start Discord session", err)

	s.AddHandler(handleInteractionCreate)
	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Default().Printf("[Discord] %s is ready", s.State.User)
	})

	err = s.Open()
	util.CheckError("Failed to open Discord session", err)

	return s
}

func handleInteractionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.Type {
	case discordgo.InteractionMessageComponent:
		buttonID := i.MessageComponentData().CustomID

		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseModal,
			Data: &discordgo.InteractionResponseData{
				CustomID: "folder_name_modal_" + buttonID,
				Title:    "Set the folder name",
				Components: []discordgo.MessageComponent{
					discordgo.ActionsRow{
						Components: []discordgo.MessageComponent{
							discordgo.TextInput{
								CustomID:    "folder_name_input_" + buttonID,
								Label:       "Enter the folder name",
								Style:       discordgo.TextInputShort,
								Placeholder: "May 2024",
								Required:    true,
								MaxLength:   20,
								MinLength:   1,
							},
						},
					},
				},
			},
		})
		util.CheckError("Failed to send modal", err)
	case discordgo.InteractionModalSubmit:
		data := i.ModalSubmitData()
		folderName := data.Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value

		after, _ := strings.CutPrefix(data.CustomID, "folder_name_modal_")
		urls := strings.Split(after, ",")
		// TODO: set folder names
		log.Default().Println(urls)

		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{
					{
						Title:       "Folder name set!",
						Description: folderName,
						Color:       0x32FF25,
					},
				},
			},
		})
		util.CheckError("Failed to respond to modal submission", err)
	}
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
					discordgo.Button{
						Label:    "Set folder name",
						Style:    discordgo.PrimaryButton,
						CustomID: fmt.Sprintf("%s,%s", s3Url, driveUrl),
						Emoji: &discordgo.ComponentEmoji{
							Name: "📁",
						},
					},
				},
			},
		},
	})
	util.CheckError("Failed to send success message", err)
}
