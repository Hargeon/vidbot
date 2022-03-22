// Package chain uses for processing user messages
package chain

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/Hargeon/vidbot/pkg/service"
	"github.com/Hargeon/vidbot/pkg/service/videocmprs"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const expTime = time.Hour

var pollQuestions = []string{"Resolution", "Aspect Ratio", "Bitrate"}

type Base struct {
	// vSrv - client for videocmprs service
	VSrv  service.VideoService
	Cache service.Cache
	Bot   *tgbotapi.BotAPI

	Next service.MessageHandler
}

// SetNext chain
func (b *Base) SetNext(next service.MessageHandler) {
	b.Next = next
}

// AddToCache current state
func (b *Base) AddToCache(acc *videocmprs.Account) error {
	res, err := json.Marshal(acc)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("%d", acc.ChatID)
	fmt.Println("add to cache key", key, "value", string(res))
	return b.Cache.Write(key, res, expTime)
}

// DeleteFromCache ...
func (b *Base) DeleteFromCache(chatID int64) error {
	key := fmt.Sprintf("%d", chatID)

	return b.Cache.Delete(key)
}

func (b *Base) sendPoll(chatID int64, tokenAuth string) error {
	account := &videocmprs.Account{
		TokenAuth: tokenAuth,
		State:     videocmprs.Attributes,
		ChatID:    chatID,
	}

	if err := b.AddToCache(account); err != nil {
		return err
	}

	// send poll
	poll := tgbotapi.NewPoll(chatID, "What do you want to change", pollQuestions...)
	poll.AllowsMultipleAnswers = true
	poll.IsAnonymous = false
	b.SendMsg(poll)

	return nil
}

// SendMsg to telegram
func (b *Base) SendMsg(c tgbotapi.Chattable) {
	if _, err := b.Bot.Send(c); err != nil {
		fmt.Println("err when send msg", err)
	}
}

// SomethingWentWrong send error to telegram
func (b *Base) SomethingWentWrong(chatID int64) {
	msg := tgbotapi.NewMessage(chatID, "Somening went wrong, try later")

	b.SendMsg(msg)
}
