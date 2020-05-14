package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/slack-go/slack"

	"github.com/n0mzk/kfc-reactor/db"
	"github.com/n0mzk/kfc-reactor/handlers"
)

const (
	envSlackBotToken      = "SLACK_BOT_TOKEN"
	envSlackUserToken     = "SLACK_USER_TOKEN"
	envSlackSigningSecret = "KFC_REACTOR_SIGNING_SECRET"
	envPort               = "PORT"
	envHomeChannelID      = "KFC_REACTOR_HOME_CHANNEL_ID"
	envRedisURL           = "REDIS_URL"
)

var (
	port    string
	handler *handlers.Handler
)

func init() {
	logger := log.New(os.Stdout, "kfc-reactor: ", log.Lshortfile|log.LstdFlags)
	botToken := os.Getenv(envSlackBotToken)
	if botToken == "" {
		logger.Fatal(envSlackBotToken + " is not provided")
	}
	userToken := os.Getenv(envSlackUserToken)
	if userToken == "" {
		logger.Fatal(envSlackUserToken + " is not provided")
	}
	signingSecret := os.Getenv(envSlackSigningSecret)
	if signingSecret == "" {
		logger.Fatal(envSlackSigningSecret + " is not provided")
	}
	port = os.Getenv(envPort)
	if port == "" {
		logger.Fatal(envPort + " is not provided")
	}
	homeChannel := os.Getenv(envHomeChannelID)
	if homeChannel == "" {
		logger.Fatal(envHomeChannelID + " is not provided")
	}
	redisURL := os.Getenv(envRedisURL)
	if redisURL == "" {
		logger.Fatal(envRedisURL + " is not provided")
	}

	database, err := db.NewDB(redisURL, logger)
	if err != nil {
		logger.Fatalf("new redis connection failed: %s", err)
	}

	handler, err = handlers.NewHandler(slack.New(botToken), slack.New(userToken), signingSecret, homeChannel, logger, database)
	if err != nil {
		logger.Fatalf("new handler failed: %s", err)
	}
}

func main() {
	http.HandleFunc("/command", handler.HandleSlashCommands)
	http.HandleFunc("/", handler.HandleEvents)

	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatalf("listen and serve failed: %s", err)
	}
	fmt.Println("listening")
}
