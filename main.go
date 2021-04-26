package main

import (
	"flag"
	"chandrayaan/cmd"
	"chandrayaan/models"
	"github.com/BurntSushi/toml"
	filename "github.com/keepeye/logrus-filename"
	log "github.com/sirupsen/logrus"
	_ "net/http/pprof"
	"os"
)

func main() {

	var mode *string
	var configFile *string
	mode = flag.String("mode", "server", "server/sync")
	configFile = flag.String("config", "config.toml", "configuration file")
	flag.Parse()

	var config models.Configuration
	if _, err := toml.DecodeFile(*configFile, &config); err != nil {
		log.Fatalln("error parsing config file", err)
	}

	configureLogger(config)

	switch *mode {
	case "server":
		cmd.BotMode(config)
	case "sync":
		cmd.SyncMode(config)
	default:
		log.Fatalln("invalid mode")
	}

}

// todo: fix logging, add timestamp and filename
func configureLogger(config models.Configuration) {
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)
	filenameHook := filename.NewHook()
	filenameHook.Field = "L"
	log.AddHook(filenameHook)
	log.SetFormatter(&log.TextFormatter{
		TimestampFormat:        "2006-01-02 15:04:05",
		FullTimestamp:          true,
		ForceColors:            true,
		DisableLevelTruncation: false,
	})
}
