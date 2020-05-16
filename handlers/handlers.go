package handlers

import (
	"fmt"
	"golang.org/x/text/unicode/norm"
	"log"
	"strings"
	"unicode"

	"github.com/slack-go/slack"

	"github.com/n0mzk/kfc-reactor/db"
)

type Handler struct {
	botClient     *slack.Client
	userClient    *slack.Client
	signingSecret string
	homeChannel   string
	logger        *log.Logger
	Database      *db.DB
}

func NewHandler(botClient, userClient *slack.Client, secret, homeCh string, logger *log.Logger, database *db.DB) (*Handler, error) {
	keywords, err := database.ListKeywords()
	if err != nil {
		return nil, fmt.Errorf("get keywords list failed: %w", err)
	}
	db.Keywords = keywords
	kanameMadokas, err := database.ListKanameMadokas()
	if err != nil {
		return nil, fmt.Errorf("get kaname madokas list failed: %w", err)
	}
	db.KanameMadokas = kanameMadokas

	return &Handler{
		botClient:     botClient,
		userClient:    userClient,
		signingSecret: secret,
		homeChannel:   homeCh,
		logger:        logger,
		Database:      database,
	}, nil
}

func (h *Handler) handleErr(err error, chId, msg string) {
	h.logger.Println(err)
	h.sendMessage(chId, msg)
}

func (h *Handler) sendMessage(chId, msg string) {
	_, _, _, err := h.botClient.SendMessage(
		chId,
		slack.MsgOptionAsUser(false),
		slack.MsgOptionText(msg, false),
	)
	if err != nil {
		h.logger.Printf("send message failed: %s", err)
	}
}

func (h *Handler) isKanameMadoka(userID string) bool {
	for _, v := range db.KanameMadokas {
		if userID == v.UserId {
			return true
		}
	}
	return false
}

func (h *Handler) contains(s string) bool {
	s = h.normalize(s)
	for _, v := range db.Keywords {
		if !strings.Contains(s, v) {
			continue
		}
		return strings.Contains(s, v)
	}
	return false
}

var avoidingSearchChars = []string{
	" ",
	".",
	"/",
	"_",
	"-",
}

func (h *Handler) normalize(s string) string {
	// unicode NFKC normalize
	s = norm.NFKC.String(s)

	// upper case to lower case
	s = strings.ToLower(s)

	// Katakana to Hiragana
	hiragana := make([]rune, len([]rune(s)))
	for i, v := range []rune(s) {
		if unicode.In(v, unicode.Katakana) {
			hiragana[i] = v - 96
		} else {
			hiragana[i] = v
		}
	}
	s = string(hiragana)

	// avoid avoiding-search
	for _, c := range avoidingSearchChars {
		s = strings.Replace(s, c, "", -1)
	}

	return s
}
