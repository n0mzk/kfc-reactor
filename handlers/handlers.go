package handlers

import (
	"log"

	"github.com/slack-go/slack"
)

type Handler struct {
	BotClient     *slack.Client
	UserClient    *slack.Client
	SigningSecret string
	logger        *log.Logger
}

func NewHandler(botClient, userClient *slack.Client, secret string, logger *log.Logger) *Handler {
	return &Handler{
		BotClient:     botClient,
		UserClient:    userClient,
		SigningSecret: secret,
		logger:        logger,
	}
}
