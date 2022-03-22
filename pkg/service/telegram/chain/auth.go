package chain

import (
	"errors"

	"github.com/Hargeon/vidbot/pkg/service/videocmprs"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	notRegisteredMsg = "You are not registered."
)

type Auth struct {
	Base
}

// Execute function uses for authorization
func (a *Auth) Execute(chatID int64, acc *videocmprs.Account, update tgbotapi.Update) {
	if acc != nil {
		a.Next.Execute(chatID, acc, update)

		return
	}

	tokenAuth, err := a.VSrv.Authenticate(chatID)
	if err != nil {
		if !errors.Is(err, videocmprs.ErrNotRegistered) {
			a.SomethingWentWrong(chatID)

			return
		}

		if err := a.registerState(chatID); err != nil {
			a.SomethingWentWrong(chatID)
		}

		return
	}

	if err := a.sendPoll(chatID, tokenAuth); err != nil {
		a.SomethingWentWrong(chatID)
	}
}

func (a *Auth) registerState(chatID int64) error {
	account := &videocmprs.Account{
		State:  videocmprs.Registration,
		ChatID: chatID,
	}

	if err := a.AddToCache(account); err != nil {
		return err
	}

	msgStr := notRegisteredMsg + " Please, Send your email."
	msg := tgbotapi.NewMessage(chatID, msgStr)
	a.SendMsg(msg)

	return nil
}
