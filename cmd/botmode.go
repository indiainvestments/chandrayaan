package cmd

import (
	"chandrayaan/models"
	"chandrayaan/service/corp_action"
	"chandrayaan/service/corp_news"
	"chandrayaan/service/discordbot"
	"chandrayaan/store"
	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func BotMode(config models.Configuration) {

	// Profiler
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	db, err := configureDatabase(config)
	if err != nil {
		log.Fatalln("error opening db session", err)
	}

	dbStore := store.NewStore(db)

	// handles refreshing, storing and notifying corp news
	newsNotifier, err := corp_news.CorporateNewsNotifier(corp_news.SchedulerOpts{
		CorporateNewsFetcher:         corp_news.NewBseCorporateNewsFetcher(),
		CorporateNewsRefreshInterval: time.Minute * time.Duration(config.NewsRefreshMinutes),
		Store:                        dbStore,
		Workers:                      config.ScrapeWorkers,
	})
	if err != nil {
		log.Fatalln("error opening news scheduler session", err)
	}

	discord, err := discordgo.New(config.DiscordBotToken)
	if err != nil {
		log.Fatal("error opening discord session", err)
	}
	defer discord.Close()

	cyBot := discordbot.NewCorporateNewsBot(dbStore, discord)

	// in case we want other notifications, we will register this channel to other event notifier
	eventsChannel := make(chan interface{}, 10)
	defer close(eventsChannel)

	newsNotifier.Register(eventsChannel)
	defer newsNotifier.Unregister(eventsChannel)

	// handles refreshing, storing and notifying corp action
	actionNotifier, err := corp_action.CorporateActionNotifier(corp_action.CorporateActionOpts{
		CorporateActionFetcher:       corp_action.NewBseCorporateActionFetcher(),
		CorporateNewsRefreshInterval: time.Minute * time.Duration(config.ActionRefreshMinutes),
		Store:                        dbStore,
		Workers:                      config.ScrapeWorkers,
	})
	if err != nil {
		log.Fatalln("error opening action scheduler session", err)
	}

	actionNotifier.Register(eventsChannel)
	defer actionNotifier.Unregister(eventsChannel)

	// todo: use another struct to manage this
	go func() {
		for msg := range eventsChannel {
			switch event := msg.(type) {
			case corp_news.CorporateNews:
				if err := cyBot.SendNewsNotifications(event); err != nil {
					log.Error("error sending notification to user", err)
				}
			case corp_action.CorporateAction:
				// TODO: send notif

			default:
				log.Error("unknown event type generated", event)
			}
		}
	}()

	commandCenter := discordbot.NewCommandCentre("!cy", discord)
	commandCenter.MustRegister("search", cyBot.SearchScrip)
	commandCenter.MustRegister("help", cyBot.Help)
	commandCenter.MustRegister("corp-news", cyBot.CoporateNews)
	commandCenter.MustRegister("corp-action", cyBot.CorporateAction)
	commandCenter.MustRegister("notify", cyBot.Notifications)
	commandCenter.MustRegister("ticker-subscription", cyBot.TickerSubscription)
	commandCenter.Start()

	killSignal := make(chan os.Signal, 1)
	signal.Notify(killSignal, os.Interrupt)
	<-killSignal
}

func configureDatabase(config models.Configuration) (*gorm.DB, error) {
	var dbLogger = logger.New(
		log.New(),
		logger.Config{
			SlowThreshold: time.Second * 10,
			LogLevel:      logger.Info,
			Colorful:      false,
		},
	)

	if !config.LogQuery {
		dbLogger = logger.Default.LogMode(logger.Silent)
	}

	db, err := gorm.Open(postgres.Open(config.DSN), &gorm.Config{
		Logger: dbLogger,
	})
	if err != nil {
		return nil, err
	}

	rawDb, err := db.DB()
	if err != nil {
		return nil, err
	}

	// SetMaxIdleConns sets the maximum number of connections in the idle connection pool.
	rawDb.SetMaxIdleConns(50)

	// SetMaxOpenConns sets the maximum number of open connections to the database.
	rawDb.SetMaxOpenConns(100)

	// SetConnMaxLifetime sets the maximum amount of time a connection may be reused.
	rawDb.SetConnMaxLifetime(time.Hour)

	return db, nil
}
