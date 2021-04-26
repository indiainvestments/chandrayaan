package corp_action

import (
	"chandrayaan/service/events_notifier"
	"chandrayaan/service/pubsub"
	"chandrayaan/store"
	"errors"
	log "github.com/sirupsen/logrus"
	"time"
)

type CorporateActionOpts struct {
	CorporateActionFetcher       *BseCorporateActionFetcher
	Workers                      int
	CorporateNewsRefreshInterval time.Duration
	Store                        *store.Store
}

type corporateActionWorkerInput struct {
	store.TickerInfo
}

type corporateActionWorkerOutput CorporateAction

type ActionNotifier struct {
	refreshInterval        time.Duration
	store                  *store.Store
	corporateActionFetcher *BseCorporateActionFetcher
	corporateNewsWorkers   int

	workerInput  chan corporateActionWorkerInput
	workerOutput chan corporateActionWorkerOutput

	pubsub *pubsub.PubSub
}

func (s *ActionNotifier) corporateActionWorker() {

	for newsQuery := range s.workerInput {
		allNews, err := s.corporateActionFetcher.GetNews(newsQuery.TickerInfo)
		if err != nil {
			log.Error(err)
			continue
		}
		s.workerOutput <- corporateActionWorkerOutput(*allNews)
	}

}

func (s *ActionNotifier) corporateActionCollector() {
	for corpActions := range s.workerOutput {

		var action = CorporateAction{
			Ticker: corpActions.Ticker,
		}

		for _, i := range corpActions.Actions {

			exDate, _ := time.Parse("02 Jan 2006", i.ExDate)
			paymentDate, _ := time.Parse("2006-01-02T15:04:05", i.PaymentDate)

			_, err := s.store.CorporateActionStore.Add(&store.CorporateAction{
				ExDate:      &exDate,
				Purpose:     i.Purpose,
				Details:     i.Details,
				PaymentDate: &paymentDate,
				Ticker:      corpActions.Ticker,
			})
			if err == nil {
				action.Actions = append(action.Actions, i)
			} else if err != store.ErrDuplicate {
				log.Error(err)
			}
		}

		if len(action.Actions) > 0 {
			s.pubsub.Broadcast(action)
		}
	}
}

// Will return CorporateNews in channel
func (s *ActionNotifier) Register(c chan interface{}) {
	s.pubsub.Register(c)
}

func (s *ActionNotifier) Unregister(c chan interface{}) {
	s.pubsub.UnRegister(c)
}

func (s *ActionNotifier) refreshCorporateAction() {

	for range time.Tick(s.refreshInterval) {
		iterableScrip, err := s.store.BseTickerStore.List()
		if err != nil {
			log.Error(err)
			continue
		}

		for _, i := range iterableScrip {
			s.workerInput <- corporateActionWorkerInput{
				TickerInfo: i,
			}
		}
	}
}

func CorporateActionNotifier(opts CorporateActionOpts) (events_notifier.EventsNotifier, error) {

	if opts.CorporateActionFetcher == nil {
		return nil, errors.New("unexpected value for CorporateNewsFetcher")
	}

	if opts.Store == nil {
		return nil, errors.New("unexpected value for Store")
	}

	if opts.CorporateNewsRefreshInterval == 0 {
		opts.CorporateNewsRefreshInterval = time.Hour * 24
	}

	if opts.Workers <= 0 {
		opts.Workers = 10
	}

	pubSub := pubsub.NewPubSub()
	pubSub.Start()

	s := &ActionNotifier{
		corporateActionFetcher: opts.CorporateActionFetcher,
		refreshInterval:        opts.CorporateNewsRefreshInterval,
		store:                  opts.Store,
		corporateNewsWorkers:   opts.Workers,
		workerInput:            make(chan corporateActionWorkerInput, opts.Workers),
		workerOutput:           make(chan corporateActionWorkerOutput, opts.Workers),
		pubsub:                 pubSub,
	}

	for i := 0; i < s.corporateNewsWorkers; i++ {
		go s.corporateActionWorker()
		go s.corporateActionCollector()
	}

	go s.refreshCorporateAction()

	return s, nil
}
