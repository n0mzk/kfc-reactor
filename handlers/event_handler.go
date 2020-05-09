package handlers

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"

	"github.com/n0mzk/kfc-reactor/config"
)

func (h *Handler) HandleEvents(w http.ResponseWriter, r *http.Request) {
	h.logger.Println("event received")
	verifier, err := slack.NewSecretsVerifier(r.Header, h.SigningSecret)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.logger.Printf("new secrets verifier failed: %s", err)
		return
	}
	reader := io.TeeReader(r.Body, &verifier)
	body, err := ioutil.ReadAll(reader)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.logger.Printf("read body failed: %s", err)
		return
	}
	if err := verifier.Ensure(); err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		h.logger.Printf("verification failed: %s", err)
		return
	}

	ev, err := slackevents.ParseEvent(json.RawMessage(body), slackevents.OptionNoVerifyToken())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.logger.Print(err)
		return
	}

	switch ev.Type {
	case slackevents.URLVerification:
		var res *slackevents.ChallengeResponse
		if err := json.Unmarshal([]byte(body), &res); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			h.logger.Printf("unmarshal body failed: %s", err)
			return
		}
		w.Header().Set("Content-Type", "text/plain")
		if _, err := w.Write([]byte(res.Challenge)); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			h.logger.Printf("write response failed: %s", err)
			return
		}
	case slackevents.CallbackEvent:
		switch typed := ev.InnerEvent.Data.(type) {
		case *slackevents.MessageEvent:
			if config.Contains(typed.Text) {
				ref := slack.ItemRef{
					Channel:   typed.Channel,
					Timestamp: typed.TimeStamp,
				}
				err = h.UserClient.AddReaction("kfc", ref)
				if err != nil {
					h.logger.Print(err)
				}
				h.logger.Println("reaction added")
			}
		}
	}
}
