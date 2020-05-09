package config

import (
	"log"
	"strings"
	"time"

	"github.com/jinzhu/configor"
	"github.com/slack-go/slack"
)

type Config struct {
	Keywords []string `yaml:"keywords"`
}

type ConfigLoader struct {
	logger          *log.Logger
	botClient       *slack.Client
	ownersChannelID string
}

var (
	Keywords    []string
	KeywordsNum int
)

func NewConfigLoader(logger *log.Logger, botClient *slack.Client, ownerChID string) *ConfigLoader {
	return &ConfigLoader{
		logger:          logger,
		botClient:       botClient,
		ownersChannelID: ownerChID,
	}
}

func (cl *ConfigLoader) LoadConfig() {
	var c Config
	configor.New(&configor.Config{
		AutoReload:         true,
		AutoReloadInterval: time.Minute,
		AutoReloadCallback: cl.callback,
	}).Load(&c, "keywords.yml")

	Keywords = c.Keywords
	KeywordsNum = len(c.Keywords)
	cl.logger.Printf("%d keywords are loaded", KeywordsNum)
}

func (cl *ConfigLoader) callback(config interface{}) {
	c, ok := config.(*Config)
	if !ok {
		cl.logger.Println("load config failed")
	}
	if diff := len(c.Keywords) - KeywordsNum; diff > 0 {
		Keywords = c.Keywords
		KeywordsNum = len(Keywords)

		added := ""
		for _, v := range c.Keywords[len(c.Keywords)-diff:] {
			added = added + v
		}
		cl.logger.Printf("%s is added", added)

		_, _, _, err := cl.botClient.SendMessage(
			cl.ownersChannelID,
			slack.MsgOptionAsUser(false),
			slack.MsgOptionText(added+"にも反応できるようになりました！", false),
		)
		if err != nil {
			cl.logger.Printf("send message failed: %s", err)
		}
	}
}

func Contains(s string) bool {
	if strings.Count(s, "") == 4 && strings.HasPrefix(s, "ひ") && strings.HasSuffix(s, "る") {
		return true
	}
	for _, v := range Keywords {
		if !strings.Contains(s, v) {
			continue
		}
		return strings.Contains(s, v)
	}
	return false
}
