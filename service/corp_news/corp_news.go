package corp_news

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"time"
)



type CorporateNewsOpts struct {
	From *time.Time
	To   *time.Time
}

type CorporateNewsItem struct {
	Attachment string `json:"attachment,omitempty"`
	Headline   string `json:"headline,omitempty"`
	Date       string `json:"date,omitempty"`
	Category   string `json:"category,omitempty"`
	Id         string `json:"id,omitempty"`
	NewsSub    string `json:"news_sub,omitempty"`
}

type CorporateNews struct {
	News   []CorporateNewsItem
	Ticker string
}

func (n CorporateNews) ToDiscordMessage() *discordgo.MessageEmbed {

	var fields []*discordgo.MessageEmbedField

	for _, result := range n.News {

		date := result.Date
		parsedTime, err := time.Parse("2006-01-02T15:04:05.999Z", result.Date)
		if err == nil {
			date = parsedTime.Format("02 Jan 06 15:04")
		}

		headLine := result.Headline
		if len(headLine) == 0 {
			headLine = result.NewsSub
		}

		value := fmt.Sprintf("```%s\n\n%s\n\n```", headLine, date)
		if len(result.Attachment) > 0 {
			value = fmt.Sprintf("%s\n[Attachment](%s)\n\n", value, result.Attachment)
		}

		name := result.Category
		if len(result.Category) == 0 {
			name = "Company Update"
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
