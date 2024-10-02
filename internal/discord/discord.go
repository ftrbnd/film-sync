package discord

import (
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/ftrbnd/film-sync/internal/aws"
	"github.com/ftrbnd/film-sync/internal/google"
	"github.com/ftrbnd/film-sync/internal/util"
)

var bot *discordgo.Session

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
		util.CheckError("Failed to send modal", err)
	case discordgo.InteractionModalSubmit:
		data := i.ModalSubmitData()
		folderName := data.Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value

		after, _ := strings.CutPrefix(data.CustomID, "folder_name_modal_")
		ids := strings.Split(after, ",")
		go aws.SetFolderName(ids[0], folderName)
		go google.SetFolderName(ids[1], folderName)

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
	if err != nil {
		return err
	}

	log.Default().Println("Waiting for user to authenticate...")
	return nil
}

func SendSuccessMessage(s3Folder string, driveFolderID string, message string) error {
	channel, err := createDMChannel()
	if err != nil {
		return err
	}

	region, err := util.LoadEnvVar("AWS_REGION")
	if err != nil {
		return err
	}
	bucket, err := util.LoadEnvVar("AWS_BUCKET_NAME")
	if err != nil {
		return err
	}

	s3Url := fmt.Sprintf("https://%s.console.aws.amazon.com/s3/buckets/%s?region=%s&prefix=%s/", region, bucket, region, s3Folder)
	driveUrl := fmt.Sprintf("https://drive.google.com/drive/u/0/folders/%s", driveFolderID)

	_, err = bot.ChannelMessageSendComplex(channel.ID, &discordgo.MessageSend{
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
