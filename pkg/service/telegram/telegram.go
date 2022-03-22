// Package telegram uses for reading messages from telegram
package telegram

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime/debug"
	"time"

	"github.com/Hargeon/vidbot/pkg/service"
	"github.com/Hargeon/vidbot/pkg/service/telegram/chain"
	"github.com/Hargeon/vidbot/pkg/service/videocmprs"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	reConTimeout = time.Second * 2
	reconTimes   = 5
)

type Bot struct {
	BotClient *tgbotapi.BotAPI
	cache     service.Cache
	// vSrv - client for videocmprs service
	vSrv service.VideoService

	// endpoint to telegram bot
	botEnd string
	token  string
}

// NewBot initialzie Bot
func NewBot(token string, c service.Cache, vSrv service.VideoService) (*Bot, error) {
	tgHost := os.Getenv("TELEGRAM_HOST")
	botEnd := tgHost + "/bot%s/%s"

	bb := &Bot{cache: c, botEnd: botEnd, token: token, vSrv: vSrv}

	bot, err := reconectBot(token, botEnd)

	bb.BotClient = bot

	return bb, err
}

func reconectBot(token, botEnd string) (*tgbotapi.BotAPI, error) {
	var tm int
	var bot *tgbotapi.BotAPI
	var err error
	for i := 0; i < reconTimes; i++ {
		bot, err = tgbotapi.NewBotAPIWithAPIEndpoint(token, botEnd)

		if err == nil {
			break
		}

		tm++
		time.Sleep(reConTimeout)
	}

	return bot, err
}

// ReadMessages from telegram
func (b *Bot) ReadMessages() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.BotClient.GetUpdatesChan(u)

	wr := &chain.Wrong{
		Base: chain.Base{
			Bot: b.BotClient,
		},
	}

	video := &chain.Video{
		Base: chain.Base{
			VSrv:  b.vSrv,
			Cache: b.cache,
			Bot:   b.BotClient,
		},
	}

	video.SetNext(wr)

	attr := &chain.Attribute{
		Base: chain.Base{
			VSrv:  b.vSrv,
			Cache: b.cache,
			Bot:   b.BotClient,
		},
	}

	attr.SetNext(video)

	poll := &chain.Poll{
		Base: chain.Base{
			VSrv:  b.vSrv,
			Cache: b.cache,
			Bot:   b.BotClient,
		},
	}

	poll.SetNext(attr)

	reqistr := &chain.Registration{
		Base: chain.Base{
			VSrv:  b.vSrv,
			Cache: b.cache,
			Bot:   b.BotClient,
		},
	}

	reqistr.SetNext(poll)

	auth := &chain.Auth{
		Base: chain.Base{
			VSrv:  b.vSrv,
			Cache: b.cache,
			Bot:   b.BotClient,
		},
	}

	auth.SetNext(reqistr)

	for u := range updates {
		go b.processMsg(auth, u)
	}
}

func (b *Bot) processMsg(chain service.MessageHandler, update tgbotapi.Update) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("err recover", err)
			fmt.Println("stacktrace from panic: ", string(debug.Stack()))
		}
	}()

	var chatID int64
	if update.Message != nil {
		chatID = update.Message.Chat.ID
	} else if update.PollAnswer != nil {
		chatID = update.PollAnswer.User.ID
	} else {
		fmt.Println("failed")
		return
	}

	acc, err := b.readFromCache(chatID)
	if err != nil {
		acc = nil
	}

	chain.Execute(chatID, acc, update)
}

func (b *Bot) readFromCache(chatID int64) (*videocmprs.Account, error) {
	key := fmt.Sprintf("%d", chatID)
	resStr, err := b.cache.Read(key)
	if err != nil {
		return nil, err
	}

	fmt.Println("read cache key", key, "value", resStr)

	res := new(videocmprs.Account)
	err = json.Unmarshal([]byte(resStr), res)

	return res, err
}
