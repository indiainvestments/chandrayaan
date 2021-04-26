package corp_action

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
)

type CorporateActionItem struct {
	ExDate      string `json:"ex_date"`
	Purpose     string `json:"purpose"`
	Details     string `json:"details"`
	PaymentDate string `json:"payment_date"`
}

type CorporateAction struct {
	Ticker  string
	Actions []CorporateActionItem
}

func (a CorporateAction) ToDiscordMessage() *discordgo.MessageEmbed {

	var fields []*discordgo.MessageEmbedField

	for _, result := range a.Actions {

		value := "```"

		if len(result.Details) > 0 {
			value += fmt.Sprintf("Value: %s\n", result.Details)
		}

		if len(result.PaymentDate) > 0 {
			value += fmt.Sprintf("Payment Date: %s\n", result.PaymentDate)
		}
		if len(result.ExDate) > 0 {
			value += fmt.Sprintf("Ex Date: %s\n", result.ExDate)
		}

		value += "```"

		name := result.Purpose
		if len(result.Purpose) == 0 {
			name = "Action"
		}

		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   name,
			Value:  value,
			Inline: false,
		})
	}

	return &discordgo.MessageEmbed{
		Type:   discordgo.EmbedTypeRich,
		Color:  111111,
		Fields: fields,
	}

}
