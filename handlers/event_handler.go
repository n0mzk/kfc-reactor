package handlers

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"

	"github.com/n0mzk/kfc-reactor/db"
)

func (h *Handler) HandleEvents(w http.ResponseWriter, r *http.Request) {
	h.logger.Println("event received")
	verifier, err := slack.NewSecretsVerifier(r.Header, h.signingSecret)
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
			if h.isKanameMadoka(typed.User) || typed.Channel == h.homeChannel {
				return
			}
			typ, keyword := h.contains(typed.Text)
			switch {
			case typ == "KFC":
				ref := slack.ItemRef{
					Channel:   typed.Channel,
					Timestamp: typed.TimeStamp,
				}
				if err = h.userClient.AddReaction("551", ref); err != nil {
					h.logger.Print(err)
				}
				h.logger.Println("reaction added")
				link, err := h.userClient.GetPermalink(
					&slack.PermalinkParameters{
						Channel: typed.Channel,
						Ts:      typed.TimeStamp,
					},
				)
				if err != nil {
					h.logger.Printf("get message parmalink failed: %s", err)
				}
				msg := "message contains " + keyword + "\n" + link
				h.sendMessage(h.homeChannel, msg)
			case typ == "Ice":
				ref := slack.ItemRef{
					Channel:   typed.Channel,
					Timestamp: typed.TimeStamp,
				}
				if err = h.userClient.AddReaction("ice_ha_chigau", ref); err != nil {
					h.logger.Print(err)
				}
				if err = h.userClient.AddReaction("nja_naikana", ref); err != nil {
					h.logger.Print(err)
				}
				h.logger.Println("reaction added")
				link, err := h.userClient.GetPermalink(
					&slack.PermalinkParameters{
						Channel: typed.Channel,
						Ts:      typed.TimeStamp,
					},
				)
				if err != nil {
					h.logger.Printf("get message parmalink failed: %s", err)
				}
				msg := "message contains " + keyword + "\n" + link
				h.sendMessage(h.homeChannel, msg)
			case typ == "":
				// do nothing
			}
		case *slackevents.ReactionAddedEvent:
			for _, v := range db.KanameMadokas {
				if typed.ItemUser == v.UserId {
					h.userClient.RemoveReaction(typed.Reaction, slack.ItemRef{
						Channel:   typed.Item.Channel,
						Timestamp: typed.EventTimestamp,
					})
				}
				h.logger.Println("reaction removed")
			}
		}
	}
}
