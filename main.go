package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

const (
	envSlackBotToken      = "SLACK_BOT_TOKEN"
	envSlackUserToken     = "SLACK_USER_TOKEN"
	envSlackSigningSecret = "KFC_REACTOR_SIGNING_SECRET"
	envPort               = "PORT"
)

func main() {
	botToken := os.Getenv(envSlackBotToken)
	if botToken == "" {
		log.Fatal(envSlackBotToken + "is not provided")
	}
	userToken := os.Getenv(envSlackUserToken)
	if userToken == "" {
		log.Fatal(envSlackUserToken + "is not provided")
	}
	secret := os.Getenv(envSlackSigningSecret)
	if secret == "" {
		log.Fatal(envSlackSigningSecret + "is not provided")
	}
	port := os.Getenv(envPort)
	if port == "" {
		log.Fatal(envPort + "is not provided")
	}

	userClient := slack.New(userToken)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("received")
		verifier, err := slack.NewSecretsVerifier(r.Header, secret)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Fatalf("new secrets verifier failed: %s", err)
		}
		reader := io.TeeReader(r.Body, &verifier)
		body, err := ioutil.ReadAll(reader)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Fatalf("read body failed: %s", err)
		}
		if err := verifier.Ensure(); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			log.Printf("verification failed: %s", err)
		}

		ev, err := slackevents.ParseEvent(json.RawMessage(body), slackevents.OptionNoVerifyToken())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Fatal(err)
		}
		log.Println("parsed")

		switch ev.Type {
		case slackevents.URLVerification:
			var res *slackevents.ChallengeResponse
			if err := json.Unmarshal([]byte(body), &res); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				log.Fatalf("unmarshal body failed: %s", err)
			}
			w.Header().Set("Content-Type", "text/plain")
			if _, err := w.Write([]byte(res.Challenge)); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				log.Fatalf("write response failed: %s", err)
			}
		case slackevents.CallbackEvent:
			log.Println("callback event received")
			switch typed := ev.InnerEvent.Data.(type) {
			case *slackevents.MessageEvent:
				if contains(typed.Text, keywords) {
					ref := slack.ItemRef{
						Channel:   typed.Channel,
						Timestamp: typed.TimeStamp,
					}
					err = userClient.AddReaction("kfc", ref)
					if err != nil {
						log.Fatal(err)
					}
					log.Println("reaction added")
				}
			}
		}
	})
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatalf("listen and serve failed: %s", err)
	}
	fmt.Println("listening")
}

func contains(msg string, s []string) bool {
	if strings.Count(msg, "") == 4 && strings.HasPrefix(msg, "ひ") && strings.HasSuffix(msg, "る") {
		return true
	}
	for _, v := range s {
		if !strings.Contains(msg, v) {
			continue
		}
		return strings.Contains(msg, v)
	}
	return false
}
