package models

type Configuration struct {
	DSN                  string `toml:"dsn"`
	LogQuery             bool   `toml:"log_query"`
	DiscordBotToken      string `toml:"discord_bot_token"`
	NewsRefreshMinutes   int    `toml:"news_refresh_minutes"`
	ActionRefreshMinutes int    `toml:"action_refresh_minutes"`
	ScrapeWorkers        int    `toml:"workers"`
}
