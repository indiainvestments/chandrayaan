package cmd

import (
	"chandrayaan/models"
	"chandrayaan/service/corp_action"
	"chandrayaan/service/corp_news"
	"chandrayaan/store"
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"time"
)

func SyncMode(config models.Configuration) {

	db, err := gorm.Open(postgres.Open(config.DSN), &gorm.Config{})
	if err != nil {
		log.Fatalln("error opening database", err)
	}

	dbStore := store.NewStore(db)

	allScrip, err := dbStore.BseTickerStore.List()
	if err != nil {
		log.Fatalln("error getting ticker list", err)
	}

	syncNews(allScrip, dbStore)

	syncActions(allScrip, dbStore)

}

func syncNews(tickers []store.TickerInfo, dbStore *store.Store) {

	today := time.Now()
	fewDaysAgo := today.AddDate(0, 0, -90)

	fetcher := corp_news.NewBseCorporateNewsFetcher()

	for _, scrip := range tickers {

		log.Debug("Syncing News for ", scrip.Name+" ", scrip.Ticker)
		news, err := fetcher.GetNews(scrip, corp_news.CorporateNewsOpts{
			From: &fewDaysAgo,
			To:   &today,
		})
		if err != nil {
			log.Errorln(err)
			continue
		}

		for _, i := range news.News {
			_, err = dbStore.CorporateNewsStore.Add(&store.CorporateNews{
				Attachment: i.Attachment,
				Headline:   i.Headline,
				Date:       i.Date,
				Category:   i.Category,
				Id:         i.Id,
				Ticker:     news.Ticker,
				NewsSub:    i.NewsSub,
			})
			if err != nil && err != store.ErrDuplicate {
				log.Errorln(err)
				continue
			}
		}
	}

}

func syncActions(tickers []store.TickerInfo, dbStore *store.Store) {
	actionFetcher := corp_action.NewBseCorporateActionFetcher()

	for _, scrip := range tickers {

		log.Debug("Syncing Action for ", scrip.Name+" ", scrip.Ticker)

		news, err := actionFetcher.GetNews(scrip)
		if err != nil {
			log.Errorln(err)
			continue
		}

		for _, i := range news.Actions {

			exDate, _ := time.Parse("02 Jan 2006", i.ExDate)
			paymentDate, _ := time.Parse("2006-01-02T15:04:05", i.PaymentDate)

			_, err = dbStore.CorporateActionStore.Add(&store.CorporateAction{
				ExDate:      &exDate,
				Purpose:     i.Purpose,
				Details:     i.Details,
				PaymentDate: &paymentDate,
				Ticker:      news.Ticker,
			})
			if err != nil && err != store.ErrDuplicate {
				log.Errorln(err)
				continue
			}
		}

	}

}
