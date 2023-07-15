package discord

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/WiggidyW/eve-contract-notifications-go/contracts"
)

const (
	TOKEN      string = "##DISCORD_TOKEN##"
	CHANNEL_ID string = "##DISCORD_CHANNEL_ID##"
)

type NotificationContracts struct {
	New      *[]contracts.Contract
	Expiring *[]contracts.Contract
}

func AsEveTime(t time.Time) string {
	return t.Format("02-01 15:04 UTC")
}

func OpenDiscordSession() (*discordgo.Session, error) {
	discord, err := discordgo.New("Bot " + TOKEN)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to create discord session: %v",
			err,
		)
	}

	err = discord.Open()
	if err != nil {
		return nil, fmt.Errorf(
			"failed to open discord session: %v",
			err,
		)
	}

	return discord, nil
}

func WriteContracts(
	discord *discordgo.Session,
	contracts NotificationContracts,
) error {
	message := ""
	if contracts.New != nil {
		newContracts := *contracts.New
		if len(newContracts) > 0 {
			message += "New Contracts:\n"
			for _, contract := range *contracts.New {
				message += fmt.Sprintf(
					"  hash: %s issued: %s expires: %s\n",
					contract.HashCode,
					AsEveTime(contract.Issued),
					AsEveTime(contract.Expires),
				)
			}
		}
	}
	if contracts.Expiring != nil {
		expiringContracts := *contracts.Expiring
		if len(expiringContracts) > 0 {
			message += "Expiring Contracts:\n"
			for _, contract := range *contracts.Expiring {
				message += fmt.Sprintf(
					"  hash: %s issued: %s expires: %s\n",
					contract.HashCode,
					AsEveTime(contract.Issued),
					AsEveTime(contract.Expires),
				)
			}
		}
	}
	if message != "" {
		return WriteDiscordMessage(discord, message)
	} else {
		return nil
	}
}

func WriteDiscordMessage(
	discord *discordgo.Session,
	message string,
) error {
	_, err := discord.ChannelMessageSend(CHANNEL_ID, message)
	if err != nil {
		return fmt.Errorf(
			"failed to send discord message: %v",
			err,
		)
	}
	return nil
}
