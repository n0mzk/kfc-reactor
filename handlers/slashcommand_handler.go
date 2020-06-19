package handlers

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/slack-go/slack"

	"github.com/n0mzk/kfc-reactor/db"
)

const (
	addCommand          = "add"
	kanameMadokaCommand = "全てのemojiを、生まれる前に消し去りたい。全ての宇宙、過去と未来の全てのemojiを、この手で"
)

func (h *Handler) HandleSlashCommands(w http.ResponseWriter, r *http.Request) {
	h.logger.Println("slash command received")
	verifier, err := slack.NewSecretsVerifier(r.Header, h.signingSecret)
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

	if err = verifier.Ensure(); err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		h.logger.Printf("verification failed: %s", err)
		return
	}

	spl := strings.Split(cmd.Text, " ")

	switch spl[0] {
	case addCommand:
		if len(spl) != 2 {
			h.sendMessage(cmd.ChannelID, "`/kfc-reactor add [keyword]` の形式で送ってください！")
			return
		}
		typ, _ := h.contains(spl[1])
		if typ != "" {
			h.sendMessage(cmd.ChannelID, spl[1]+" にはもう反応できます！")
		} else {
			err = h.Database.AddKeyword(spl[1])
			if err != nil {
				h.handleErr(err, cmd.ChannelID, "なにかうまくいきませんでした。もう一度試してください！")
				return
			}
			h.sendMessage(cmd.ChannelID, spl[1]+" を覚えました！ありがとう！")
		}

	case kanameMadokaCommand:
		err := h.Database.AddKanameMadoka(db.KanameMadoka{UserId: cmd.UserID, UserName: cmd.UserName})
		if err != nil {
			h.handleErr(err, cmd.ChannelID, "なにかうまくいきませんでした。もう一度試してください！")
		}
		h.sendMessage(h.homeChannel, fmt.Sprintf(cmd.UserName+"は行ってしまったわ…… 円環の理に導かれて"))

	default:
		h.logger.Printf("received unknown command: %s", cmd.Text)
		h.sendMessage(cmd.ChannelID, "コマンドの内容がおかしいようです！")
	}
}
