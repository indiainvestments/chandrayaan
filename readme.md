# Chandrayaan  🚀 🌑  Discord Bot

# Configuration

```toml
# postgres data source
dsn = "host=localhost user=db_user password=db_pass dbname=db_name sslmode=disable TimeZone=Asia/Kolkata"
# log query to debug
log_query = false
# discord bot token
discord_bot_token = "Bot DISCORD_TOKEN"
# interval to scrape news
news_refresh_minutes = 30 # half hour
# interval to scrape corp action
action_refresh_minutes = 86400 # 24 hours
# concurrency to scrape
workers = 2

```

# Arguments

- `-mode`  can be either `server` where it starts the bot or `sync` where database is populated in case of fresh deployment

# Building

Need Go 1.14+

- Clone the repository.
- Download dependencies `go mod download`
- Run `./chandrayaan -config=path/to/config/yml -mode=server`

# Notes

- Sample configuration is provided [here](res/config.sample.toml)
- Postgresql schema is provided [here](res/schema.sql)