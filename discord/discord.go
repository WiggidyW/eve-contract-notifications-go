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

	CHAR_LIMIT  int    = 2000
	TITLE_EXTRA string = "```\nhash                issued (EVE)      expires (EVE)\n"
	CLOSER      string = "```"
)

type NotificationContracts struct {
	New      *[]contracts.Contract
	Expiring *[]contracts.Contract
}

func AsEveTime(t time.Time) string {
	return t.UTC().Format("01/02/06 15:04")
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
	newContractMessages := ContractsToMessages(
		"***New Contracts***\n",
		contracts.New,
	)
	expiringContractMessages := ContractsToMessages(
		"***Expiring Contracts***\n",
		contracts.Expiring,
	)
	if newContractMessages != nil {
		for _, message := range *newContractMessages {
			err := WriteDiscordMessage(discord, message)
			if err != nil {
				return err
			}
		}
	}
	if expiringContractMessages != nil {
		for _, message := range *expiringContractMessages {
			err := WriteDiscordMessage(discord, message)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func ContractsToMessages(
	titleLine string,
	contracts *[]contracts.Contract,
) *[]string {
	// If there are no contracts, return nil
	if contracts == nil {
		return nil
	}
	c := *contracts
	if len(c) == 0 {
		return nil
	}

	messages := make([]string, 0, 1)
	message := titleLine + TITLE_EXTRA
	for _, contract := range c {
		// This is the line for the contract
		hashStr := contract.HashCode
		if len(hashStr) == 15 {
			hashStr += " "
		}
		line := fmt.Sprintf(
			"%s    %s    %s\n",
			hashStr,
			AsEveTime(contract.Issued),
			AsEveTime(contract.Expires),
		)
		// If the line is impossible to fit into a message, skip it.
		if len(line)+len(titleLine)+len(TITLE_EXTRA)+len(CLOSER) > CHAR_LIMIT {
			continue
		}
		// If the line pushes the message over the limit, start a new message.
		if len(message)+len(line)+len(TITLE_EXTRA)+len(CLOSER) > CHAR_LIMIT {
			messages = append(messages, message+CLOSER)
			message = titleLine + TITLE_EXTRA
		}
		// Add the line to the message.
		message += line
	}
	// Add the last message.
	messages = append(messages, message+"```")
	return &messages
}

func WriteDiscordMessage(
	discord *discordgo.Session,
	message string,
) error {
	_, err := discord.ChannelMessageSend(
		CHANNEL_ID,
		message,
	)
	if err != nil {
		return fmt.Errorf(
			"failed to send discord message, message: '%s', err: '%v'",
			message,
			err,
		)
	}
	return nil
}
