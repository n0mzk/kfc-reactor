package handlers

import (
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/slack-go/slack"

	"github.com/n0mzk/kfc-reactor/config"
)

func (h *Handler) HandleSlashCommands(w http.ResponseWriter, r *http.Request) {
	h.logger.Println("slash command received")
	verifier, err := slack.NewSecretsVerifier(r.Header, h.SigningSecret)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.logger.Printf("new secrets verifier failed: %s", err)
		return
	}
	r.Body = ioutil.NopCloser(io.TeeReader(r.Body, &verifier))

	cmd, err := slack.SlashCommandParse(r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.logger.Printf("parse slash command failed: %s", err)
		return
	}

	if err := verifier.Ensure(); err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		h.logger.Printf("verification failed: %s", err)
		return
	}

	spl := strings.Split(cmd.Text, " ")
	if len(spl) != 2 || spl[0] != "add" {
		_, _, _, err = h.BotClient.SendMessage(
			cmd.ChannelID,
			slack.MsgOptionAsUser(false),
			slack.MsgOptionText("コマンドは `/kfc-reactor add [keyword]` の形式で送ってください :kfc:", false),
		)
		if err != nil {
			h.logger.Printf("send message failed: %s", err)
		}
		return
	}
	if config.Contains(spl[1]) {
		_, _, _, err = h.BotClient.SendMessage(
			cmd.ChannelID,
			slack.MsgOptionAsUser(false),
			slack.MsgOptionText(spl[1]+" にはもう反応できます :kfc:", false),
		)
		if err != nil {
			h.logger.Printf("send message failed: %s", err)
		}
	} else {
		yml, err := os.OpenFile("keywords.yml", os.O_APPEND|os.O_WRONLY, 0666)
		if err != nil {
			h.logger.Println("keywords.yml file not found")
			_, _, _, err = h.BotClient.SendMessage(
				cmd.ChannelID,
				slack.MsgOptionAsUser(false),
				slack.MsgOptionText("なにかうまくいきませんでした。もう一度試してください :kfc:", false),
			)
			if err != nil {
				h.logger.Printf("send message failed: %s", err)
			}
			return
		}
		_, err = yml.WriteString("\n" + `  - "` + spl[1] + `"`)
		if err != nil {
			h.logger.Printf("write to file failed: %s", err)
			_, _, _, err = h.BotClient.SendMessage(
				cmd.ChannelID,
				slack.MsgOptionAsUser(false),
				slack.MsgOptionText("なにかうまくいきませんでした。もう一度試してください :kfc:", false),
			)
			if err != nil {
				h.logger.Printf("send message failed: %s", err)
			}
			return
		}
		yml.Close()

		_, _, _, err = h.BotClient.SendMessage(
			cmd.ChannelID,
			slack.MsgOptionAsUser(false),
			slack.MsgOptionText(spl[1]+" を覚えました！ありがとう :kfc:", false),
		)
		if err != nil {
			h.logger.Printf("send message failed: %s", err)
		}
	}
}
