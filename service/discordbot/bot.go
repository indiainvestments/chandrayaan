package discordbot

import (
	"chandrayaan/service/corp_action"
	"chandrayaan/service/corp_news"
	"chandrayaan/store"
	"fmt"
	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
	"strconv"
	"strings"
	"time"
)

const DateLayout = "2006-01-02"

type CorporateNewsBot struct {
	store   *store.Store
	discord *discordgo.Session
}

func (w *CorporateNewsBot) Notifications(cmd Command) error {

	user, err := w.store.UserStore.GetOrCreate(cmd.Message.Author.ID)
	if err != nil {
		return err
	}

	switch cmd.Next() {
	case "true":
		user.Notifications = true
	case "false":
		user.Notifications = false
	}

	user, err = w.store.UserStore.Update(user)
	if err != nil {
		return err
	}

	_, _ = w.discord.ChannelMessageSendEmbed(cmd.Message.ChannelID, &discordgo.MessageEmbed{
		Type:  discordgo.EmbedTypeRich,
		Title: "Notifications for " + cmd.Message.Author.Username,
		Color: 111111,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Notifications",
				Value:  strconv.FormatBool(user.Notifications),
				Inline: false,
			},
		},
	})
	return nil

}

func (w *CorporateNewsBot) SearchScrip(cmd Command) error {

	searchTerm := cmd.Next()
	results, err := w.store.Search(searchTerm)
	if err != nil {
		return err
	}

	var Fields []*discordgo.MessageEmbedField

	for _, result := range results {
		Fields = append(Fields, &discordgo.MessageEmbedField{
			Name:   result.Ticker,
			Value:  fmt.Sprintf("```Name: %s\nIndustry: %s\n```", result.Name, result.Industry),
			Inline: false,
		})
	}

	w.discord.ChannelMessageSendEmbed(cmd.Message.ChannelID, &discordgo.MessageEmbed{
		Type:   discordgo.EmbedTypeRich,
		Title:  "Search results for " + searchTerm,
		Color:  111111,
		Fields: Fields,
	})

	return nil
}

func (w *CorporateNewsBot) TickerSubscription(cmd Command) error {

	user, err := w.store.UserStore.GetOrCreate(cmd.Message.Author.ID)
	if err != nil {
		return err
	}

	switch cmd.Next() {
	case "list":
		break
	case "add":
		for cmd.HasNext() {
			_, _ = w.store.UserScripStore.Add(store.UserScrip{
				UserId: user.Id,
				Ticker: strings.ToUpper(cmd.Next()),
			})
		}

	case "remove":
		for cmd.HasNext() {
			_ = w.store.UserScripStore.Remove(store.UserScrip{
				UserId: user.Id,
				Ticker: strings.ToUpper(cmd.Next()),
			})
		}

	default:
		return NewCommandError("no sub command provided")

	}

	scrips, err := w.store.UserScripStore.Get(user.Id)
	if err != nil {
		return err
	}

	var value = "```"
	for _, scrip := range scrips {
		value = fmt.Sprintf("%s\n%s\n", value, scrip.Ticker)
	}
	value += "```"

	if len(value) == 0 {
		value = "**No subscriptions yet**"
	}
	_, err = w.discord.ChannelMessageSendEmbed(cmd.Message.ChannelID, &discordgo.MessageEmbed{
		Type:  discordgo.EmbedTypeRich,
		Title: "Subscribed tickers for " + cmd.Message.Author.Username,
		Color: 111111,
		Fields: []*discordgo.MessageEmbedField{

			{
				Name:   "Tickers",
				Value:  value,
				Inline: false,
			},
		},
	})

	return nil

}

func (w *CorporateNewsBot) CoporateNews(cmd Command) error {

	if !cmd.HasNext() {
		return NewCommandError("no ticker provided")
	}

	var corporateNewsOpt = store.CorporateNewsGetOpt{
		Ticker: cmd.Next(),
		Limit:  5,
	}

	if cmd.HasNext() {
		corporateNewsOpt.Limit, _ = strconv.Atoi(cmd.Next())
		if corporateNewsOpt.Limit > 10 || corporateNewsOpt.Limit <= 0 {
			corporateNewsOpt.Limit = 10
		}
	}

	if cmd.HasNext() {
		parsedFrom, err := time.Parse(DateLayout, cmd.Next())
		if err != nil {
			parsedFrom = time.Now().AddDate(0, 0, -1)
		}
		corporateNewsOpt.From = &parsedFrom
	}

	if cmd.HasNext() {
		parsedTo, err := time.Parse(DateLayout, cmd.Next())
		if err != nil {
			parsedTo = time.Now()
		}
		corporateNewsOpt.To = &parsedTo
	}

	tickerNews, err := w.store.CorporateNewsStore.Get(corporateNewsOpt)

	if err != nil {
		return err
	}

	corpNews := corp_news.CorporateNews{
		Ticker: corporateNewsOpt.Ticker,
	}

	for _, i := range tickerNews {
		corpNews.News = append(corpNews.News, corp_news.CorporateNewsItem{
			Attachment: i.Attachment,
			Headline:   i.Headline,
			Date:       i.Date,
			Category:   i.Category,
			Id:         i.Id,
			NewsSub:    i.NewsSub,
		})
	}

	if len(tickerNews) == 0 {
		return NewCommandError("no news available for " + corporateNewsOpt.Ticker)
	}

	discordMessage := corpNews.ToDiscordMessage()
	discordMessage.Title = "Corporate news for " + corpNews.Ticker

	_, _ = w.discord.ChannelMessageSendEmbed(cmd.Message.ChannelID, discordMessage)

	return nil

}

func (w *CorporateNewsBot) SendNewsNotifications(latestNews corp_news.CorporateNews) error {

	subscribers, err := w.store.UserScripStore.ListSubscribers(latestNews.Ticker)

	if err != nil {
		return err
	}

	for _, userId := range subscribers {
		channel, err := w.discord.UserChannelCreate(userId)
		if err != nil {
			log.Error(err)
			continue
		}

		discordMessage := latestNews.ToDiscordMessage()
		discordMessage.Title = "Corporate News update for " + latestNews.Ticker

		w.discord.ChannelMessageSendEmbed(channel.ID, discordMessage)
	}

	return nil
}

func (w *CorporateNewsBot) Help(cmd Command) error {

	fields := []*discordgo.MessageEmbedField{
		{
			Name:   "Search for a ticker",
			Value:  "!cy search <search term>\n\n!cy search reliance",
			Inline: false,
		},
		{
			Name:   "Corporate News",
			Value:  "!cy corp-news <ticker> [count (Number)] [From (YYYY-MM-YY)] [To (YYYY-MM-YY)]\n\n!cy corp-news RELIANCE",
			Inline: false,
		},
		{
			Name:   "Corporate Action",
			Value:  "!cy corp-action <ticker>\n\n!cy corp-action ITC",
			Inline: false,
		},
		{
			Name:   "Turn notifications On or Off",
			Value:  "!cy notify {true, false}\n\n!cy notify true",
			Inline: false,
		},
		{
			Name:   "Ticker Subscription",
			Value:  "ticker-subscription {list,add,remove} [Tickers]\n\n!cy ticker-subscription list\n\n!cy ticker-subscription add RELIANCE ITC",
			Inline: false,
		},
	}

	_, _ = w.discord.ChannelMessageSendEmbed(cmd.Message.ChannelID, &discordgo.MessageEmbed{
		Type:   discordgo.EmbedTypeRich,
		Title:  "List of commands!",
		Color:  111111,
		Fields: fields,
		Footer: &discordgo.MessageEmbedFooter{
			Text: "[] is optional field(s), {} is valid options",
		},
	})

	return nil

}

func (w *CorporateNewsBot) CorporateAction(cmd Command) error {

	if !cmd.HasNext() {
		return NewCommandError("no ticker provided")
	}

	var actionOpt = store.CorporateActionGetOpt{
		Ticker: cmd.Next(),
	}

	actions, err := w.store.CorporateActionStore.Get(actionOpt)
	if err != nil {
		return err
	}

	corpNews := corp_action.CorporateAction{
		Ticker: actionOpt.Ticker,
	}

	for _, i := range actions {

		corpNews.Actions = append(corpNews.Actions, corp_action.CorporateActionItem{
			ExDate:      i.ExDate.Format("02 Jan 2006"),
			Purpose:     i.Purpose,
			Details:     i.Details,
			PaymentDate: i.PaymentDate.Format("02 Jan 2006"),
		})

	}

	discordMessage := corpNews.ToDiscordMessage()
	discordMessage.Title = "Corporate actions for " + corpNews.Ticker

	_, _ = w.discord.ChannelMessageSendEmbed(cmd.Message.ChannelID, discordMessage)

	return nil

}

func NewCorporateNewsBot(store *store.Store, discord *discordgo.Session) *CorporateNewsBot {
	return &CorporateNewsBot{
		store:   store,
		discord: discord,
	}
}
