package discord

import (
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/ftrbnd/film-sync/internal/database"
	"github.com/ftrbnd/film-sync/internal/files"
	"github.com/ftrbnd/film-sync/internal/http"
	"github.com/ftrbnd/film-sync/internal/util"
)

var bot *discordgo.Session

var dashboardURL = "https://dashboard.heroku.com/apps/film-sync"
var filmURL = "https://giosalad.dev/film"

func OpenSession() error {
	token, err := util.LoadEnvVar("DISCORD_TOKEN")
	if err != nil {
		return err
	}

	bot, err = discordgo.New("Bot " + token)
	if err != nil {
		return fmt.Errorf("unable to start discord session: %v", err)
	}

	bot.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		err := handleInteractionCreate(s, i)
		if err != nil {
			SendErrorMessage(err)
		}
	})
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

func handleInteractionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	switch i.Type {
	case discordgo.InteractionMessageComponent:
		customID := i.MessageComponentData().CustomID

		if customID == "trigger_deploy" {
			err := http.SendDeployRequest("Triggered by Discord button")
			if err != nil {
				return err
			}

			err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{
						{
							Title: "Triggered new deploy!",
							Color: 0x32FF25,
							URL:   filmURL,
						},
					},
				},
			})
			if err != nil {
				return err
			}
		} else {
			// else it is a scanID
			err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseModal,
				Data: &discordgo.InteractionResponseData{
					CustomID: "folder_name_modal_" + customID,
					Title:    "Set the folder name",
					Components: []discordgo.MessageComponent{
						discordgo.ActionsRow{
							Components: []discordgo.MessageComponent{
								discordgo.TextInput{
									CustomID:    "folder_name_input_" + customID,
									Label:       "Enter the folder name",
									Style:       discordgo.TextInputShort,
									Placeholder: "May 2024",
									Required:    true,
									MaxLength:   40,
									MinLength:   1,
								},
							},
						},
					},
				},
			})
			if err != nil {
				return err
			}
		}
	case discordgo.InteractionModalSubmit:
		data := i.ModalSubmitData()
		newFolderName := data.Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value

		scanID, _ := strings.CutPrefix(data.CustomID, "folder_name_modal_")
		scan, err := database.GetOneScan(scanID)
		if err != nil {
			return err
		}

		err = files.SetFolderNames(scan.CldFolderName, scan.DriveFolderID, newFolderName)
		if err != nil {
			return err
		}

		cldURL, driveURL, err := files.FolderLinks(scan.CldFolderName, scan.DriveFolderID)
		if err != nil {
			return err
		}

		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{
					{
						Title:       "Folder name set!",
						Description: newFolderName,
						Color:       0x32FF25,
						URL:         dashboardURL,
					},
				},
				Components: messageComponents(cldURL, driveURL, scanID),
			},
		})
		if err != nil {
			return fmt.Errorf("failed to respond to modal submission: %v", err)
		}
	}

	return nil
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
				Title:       "Authentication required",
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

func SendSuccessMessage(scanID string, message string) error {
	channel, err := createDMChannel()
	if err != nil {
		return err
	}

	scan, err := database.GetOneScan(scanID)
	if err != nil {
		return err
	}

	cldURL, driveURL, err := files.FolderLinks(scan.CldFolderName, scan.DriveFolderID)
	if err != nil {
		return err
	}

	_, err = bot.ChannelMessageSendComplex(channel.ID, &discordgo.MessageSend{
		Embeds: []*discordgo.MessageEmbed{
			{
				Title:       "Upload successful!",
				Description: message,
				Color:       0x32FF25,
				URL:         dashboardURL,
			},
		},
		Components: messageComponents(cldURL, driveURL, scanID),
	})
	if err != nil {
		return fmt.Errorf("failed to send success message: %v", err)
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

func messageComponents(cldURL string, driveURL string, scanID string) []discordgo.MessageComponent {
	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label: "Cloudinary",
					Style: discordgo.LinkButton,
					URL:   cldURL,
				}, discordgo.Button{
					Label: "Google Drive",
					Style: discordgo.LinkButton,
					URL:   driveURL,
				},
				discordgo.Button{
					Label:    "Set folder name",
					Style:    discordgo.PrimaryButton,
					CustomID: scanID,
					Emoji: &discordgo.ComponentEmoji{
						Name: "📁",
					},
				},
				discordgo.Button{
					Label:    "Trigger deploy",
					Style:    discordgo.SuccessButton,
					CustomID: "trigger_deploy",
					Emoji: &discordgo.ComponentEmoji{
						Name: "🔄",
					},
				},
			},
		},
	}

	return components
}
