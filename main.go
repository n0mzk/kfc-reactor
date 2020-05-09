package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/slack-go/slack"

	"github.com/n0mzk/kfc-reactor/config"
	"github.com/n0mzk/kfc-reactor/handlers"
)

const (
	envSlackBotToken      = "SLACK_BOT_TOKEN"
	envSlackUserToken     = "SLACK_USER_TOKEN"
	envSlackSigningSecret = "KFC_REACTOR_SIGNING_SECRET"
	envPort               = "PORT"
	envOwnersChannelID    = "KFC_REACTOR_OWNERS_CHANNEL_ID"
)

var (
	botToken      string
	userToken     string
	signingSecret string
	port          string
	ownersChannel string
	logger        *log.Logger
)

func init() {
	logger = log.New(os.Stdout, "kfc-reactor: ", log.Lshortfile|log.LstdFlags)
	botToken = os.Getenv(envSlackBotToken)
	if botToken == "" {
		logger.Fatal(envSlackBotToken + " is not provided")
	}
	userToken = os.Getenv(envSlackUserToken)
	if userToken == "" {
		logger.Fatal(envSlackUserToken + " is not provided")
	}
	signingSecret = os.Getenv(envSlackSigningSecret)
	if signingSecret == "" {
		logger.Fatal(envSlackSigningSecret + " is not provided")
	}
	port = os.Getenv(envPort)
	if port == "" {
		logger.Fatal(envPort + " is not provided")
	}
	ownersChannel = os.Getenv(envOwnersChannelID)
	if ownersChannel == "" {
		logger.Fatal(envOwnersChannelID + " is not provided")
	}

	config.NewConfigLoader(logger, slack.New(botToken), ownersChannel).LoadConfig()
}

func main() {
	h := handlers.NewHandler(slack.New(botToken), slack.New(userToken), signingSecret, logger)

	http.HandleFunc("/command", h.HandleSlashCommands)
	http.HandleFunc("/", h.HandleEvents)

	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatalf("listen and serve failed: %s", err)
	}
	fmt.Println("listening")
}
