package discord

import (
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/ftrbnd/film-sync/internal/aws"
	"github.com/ftrbnd/film-sync/internal/database"
	"github.com/ftrbnd/film-sync/internal/google"
	"github.com/ftrbnd/film-sync/internal/util"
)

var bot *discordgo.Session

var dashboardURL = "https://dashboard.heroku.com/apps/film-sync"

func OpenSession() error {
	token, err := util.LoadEnvVar("DISCORD_TOKEN")
	if err != nil {
		return err
	}

	bot, err = discordgo.New("Bot " + token)
	if err != nil {
		return fmt.Errorf("unable to start discord session: %v", err)
	}

	bot.AddHandler(handleInteractionCreate)
	bot.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Default().Printf("[Discord] %s is ready", s.State.User)
	})

	err = bot.Open()
	if err != nil {
		return fmt.Errorf("failed to open discord session: %v", err)
	}

	return nil
}

func CloseSession() error {
	return bot.Close()
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
		if err != nil {
			log.Default().Printf("Failed to send modal: %v", err)
		}
	case discordgo.InteractionModalSubmit:
		data := i.ModalSubmitData()
		folderName := data.Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value

		after, _ := strings.CutPrefix(data.CustomID, "folder_name_modal_")
		ids := strings.Split(after, ",")

		_, err := database.UpdateFolderName(ids[0], folderName)
		if err != nil {
			SendErrorMessage(err)
			return
		}
		err = google.SetFolderName(ids[1], folderName)
		if err != nil {
			SendErrorMessage(err)
			return
		}
		err = aws.SetFolderName(ids[0], folderName)
		if err != nil {
			SendErrorMessage(err)
			return
		}

		s3Url, _ := aws.FolderLink(folderName)
		driveUrl := google.FolderLink(ids[1])

		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{
					{
						Title:       "Folder name set!",
						Description: folderName,
						Color:       0x32FF25,
						URL:         dashboardURL,
					},
				},
				Components: []discordgo.MessageComponent{
					discordgo.ActionsRow{
						Components: []discordgo.MessageComponent{
							discordgo.Button{
								Label: "Drive",
								Style: discordgo.LinkButton,
								URL:   driveUrl,
							},
							discordgo.Button{
								Label: "S3",
								Style: discordgo.LinkButton,
								URL:   s3Url,
							},
							discordgo.Button{
								Label:    "Rename folder",
								Style:    discordgo.PrimaryButton,
								CustomID: fmt.Sprintf("%s,%s", folderName, ids[1]),
								Emoji: &discordgo.ComponentEmoji{
									Name: "📁",
								},
							},
						},
					},
				},
			},
		})
		if err != nil {
			log.Default().Printf("Failed to respond to modal submission: %v", err)
		}
	}
}

func createDMChannel() (*discordgo.Channel, error) {
	userID, err := util.LoadEnvVar("DISCORD_USER_ID")
	if err != nil {
		return nil, err
	}

	c, err := bot.UserChannelCreate(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to create DM channel: %v", err)
	}

	return c, nil
}

func SendAuthMessage(authURL string) error {
	channel, err := createDMChannel()
	if err != nil {
		return err
	}

	_, err = bot.ChannelMessageSendComplex(channel.ID, &discordgo.MessageSend{
		Embeds: []*discordgo.MessageEmbed{
			{
				Title:       "Authentication required!",
				Description: "Visit the link to connect with Gmail and Google Drive",
				Color:       0xFFFB25,
				URL:         dashboardURL,
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
	if err != nil {
		return err
	}

	return nil
}

func SendSuccessMessage(s3Folder string, driveFolderID string, message string) error {
	channel, err := createDMChannel()
	if err != nil {
		return err
	}

	s3Url, err := aws.FolderLink(s3Folder)
	if err != nil {
		return err
	}
	driveUrl := google.FolderLink(driveFolderID)

	_, err = bot.ChannelMessageSendComplex(channel.ID, &discordgo.MessageSend{
		Embeds: []*discordgo.MessageEmbed{
			{
				Title:       "Upload successful!",
				Description: message,
				Color:       0x32FF25,
				URL:         dashboardURL,
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
						CustomID: fmt.Sprintf("%s,%s", s3Folder, driveFolderID),
						Emoji: &discordgo.ComponentEmoji{
							Name: "📁",
						},
					},
				},
			},
		},
	})
	if err != nil {
		return err
	}

	return nil
}

func SendErrorMessage(e error) error {
	log.Default().Println(e)

	channel, err := createDMChannel()
	if err != nil {
		return err
	}

	_, err = bot.ChannelMessageSendComplex(channel.ID, &discordgo.MessageSend{
		Embeds: []*discordgo.MessageEmbed{
			{
				Title:       "Film Sync failed",
				Description: e.Error(),
				Color:       0xDF0000,
				URL:         dashboardURL,
			},
		},
	})
	if err != nil {
		return err
	}

	return nil
}
