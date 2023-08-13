package main

import (
	"collector/config"
	"collector/internal/bot_setup"
	"collector/internal/infra"
	"flag"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/sirupsen/logrus"
	"gopkg.in/telebot.v3"
	"time"
)

const configDefaultPath = "./config/config.toml"
const defaultBotPollerTimeout = 10 * time.Second

func main() {
	// Args
	var (
		configPath string
		token      string
	)
	flag.StringVar(&configPath, "c", configDefaultPath, "path to config file")
	flag.StringVar(&token, "t", "", "token for bot")
	flag.Parse()

	// Configs
	appConfig := config.NewAppConfig()
	_, err := toml.DecodeFile(configPath, &appConfig)
	if err != nil {
		logrus.Fatal(err)
	}

	// Loggers
	contextLogger := logrus.WithFields(logrus.Fields{})
	logrus.SetReportCaller(false)
	logrus.SetFormatter(&logrus.TextFormatter{PadLevelText: false, DisableLevelTruncation: false})

	// Bot
	bot, err := telebot.NewBot(telebot.Settings{
		Token:  appConfig.BotConfig.Token,
		Poller: &telebot.LongPoller{Timeout: defaultBotPollerTimeout},
	})
	if err != nil {
		logrus.Panicf("Server error: %s", fmt.Sprintf("%v", err))
	}

	// Database
	db, err := infra.NewDb(appConfig.DatabaseConfig)
	if err != nil {
		logrus.Panicf("Database error: %s", fmt.Sprintf("%v", err))
	}

	botServer := bot_setup.NewServer(contextLogger, appConfig.BotConfig.Token, db, bot)
	botServer.Run()
}
