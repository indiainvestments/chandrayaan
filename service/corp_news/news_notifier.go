package corp_news

import (
	"errors"
	"chandrayaan/service/events_notifier"
	"chandrayaan/service/pubsub"
	"chandrayaan/store"
	log "github.com/sirupsen/logrus"
	"time"
)

type SchedulerOpts struct {
	CorporateNewsFetcher         *BseCorporateNewsFetcher
	Workers                      int
	CorporateNewsRefreshInterval time.Duration
	Store                        *store.Store
}

type corporateNewsWorkerInput struct {
	store.TickerInfo
	CorporateNewsOpts
}

type corporateNewsWorkerOutput CorporateNews

type NewsNotifier struct {
	refreshInterval      time.Duration
	store                *store.Store
	corporateNewsFetcher *BseCorporateNewsFetcher
	corporateNewsWorkers int

	workerInput  chan corporateNewsWorkerInput
	workerOutput chan corporateNewsWorkerOutput

	pubsub *pubsub.PubSub
}

func (s *NewsNotifier) corporateNewsWorker() {

	for newsQuery := range s.workerInput {
		allNews, err := s.corporateNewsFetcher.GetNews(newsQuery.TickerInfo, newsQuery.CorporateNewsOpts)
		if err != nil {
			log.Error(err)
			continue
		}
		s.workerOutput <- corporateNewsWorkerOutput(*allNews)
	}

}

func (s *NewsNotifier) corporateNewsCollector() {
	for newsQuery := range s.workerOutput {

		var newNews = CorporateNews{
			Ticker: newsQuery.Ticker,
		}

		for _, i := range newsQuery.News {
			_, err := s.store.CorporateNewsStore.Add(&store.CorporateNews{
				Attachment: i.Attachment,
				Headline:   i.Headline,
				Date:       i.Date,
				Category:   i.Category,
				Id:         i.Id,
				Ticker:     newsQuery.Ticker,
				NewsSub:    i.NewsSub,
			})
			if err == nil {
				newNews.News = append(newNews.News, i)
			} else if err != store.ErrDuplicate {
				log.Error(err)
			}
		}

		if len(newNews.News) > 0 {
			s.pubsub.Broadcast(newNews)
		}
	}
}

// Will return CorporateNews in channel
func (s *NewsNotifier) Register(c chan interface{}) {
	s.pubsub.Register(c)
}

func (s *NewsNotifier) Unregister(c chan interface{}) {
	s.pubsub.UnRegister(c)
}

func (s *NewsNotifier) refreshCorporateNews() {

	for range time.Tick(s.refreshInterval) {
		now := time.Now()
		// Fetch Distinct stocks in watchlist of users

		iterableScrip, err := s.store.BseTickerStore.List()
		if err != nil {
			log.Error(err)
			continue
		}

		fewDaysAgo := now.AddDate(0, 0, -30)
		opts := CorporateNewsOpts{
			From: &fewDaysAgo,
			To:   &now,
		}
		for _, i := range iterableScrip {
			s.workerInput <- corporateNewsWorkerInput{
				TickerInfo:        i,
				CorporateNewsOpts: opts,
			}
		}
	}
}

func CorporateNewsNotifier(opts SchedulerOpts) (events_notifier.EventsNotifier, error) {

	if opts.CorporateNewsFetcher == nil {
		return nil, errors.New("unexpected value for CorporateNewsFetcher")
	}

	if opts.Store == nil {
		return nil, errors.New("unexpected value for Store")
	}

	if opts.CorporateNewsRefreshInterval == 0 {
		opts.CorporateNewsRefreshInterval = time.Hour
	}

	if opts.Workers <= 0 {
		opts.Workers = 10
	}

	pubSub := pubsub.NewPubSub()
	pubSub.Start()

	s := &NewsNotifier{
		corporateNewsFetcher: opts.CorporateNewsFetcher,
		refreshInterval:      opts.CorporateNewsRefreshInterval,
		store:                opts.Store,
		corporateNewsWorkers: opts.Workers,
		workerInput:          make(chan corporateNewsWorkerInput, opts.Workers),
		workerOutput:         make(chan corporateNewsWorkerOutput, opts.Workers),
		pubsub:               pubSub,
	}

	for i := 0; i < s.corporateNewsWorkers; i++ {
		go s.corporateNewsWorker()
		go s.corporateNewsCollector()
	}

	go s.refreshCorporateNews()

	return s, nil
}
