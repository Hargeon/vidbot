package chain

import (
	"fmt"
	"regexp"

	"github.com/Hargeon/vidbot/pkg/service/videocmprs"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const invalidEmail = "Your email is invalid. Try one more time."

var emailRegex = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

type Registration struct {
	Base
}

// Execute function uses for registration
func (r *Registration) Execute(chatID int64, acc *videocmprs.Account, update tgbotapi.Update) {
	if acc.State != videocmprs.Registration {
		r.Next.Execute(chatID, acc, update)

		return
	}

	email := update.Message.Text
	if ok := r.isEmailValid(email); !ok {
		msg := tgbotapi.NewMessage(chatID, invalidEmail)
		r.SendMsg(msg)

		return
	}

	if err := r.VSrv.Registration(chatID, email); err != nil {
		fmt.Println("req err", err)

		r.SomethingWentWrong(chatID)

		return
	}

	tokenAuth, err := r.VSrv.Authenticate(chatID)
	if err != nil {
		fmt.Println("auth err", err)

		r.SomethingWentWrong(chatID)

		return
	}

	if err := r.sendPoll(chatID, tokenAuth); err != nil {
		r.SomethingWentWrong(chatID)
	}
}

func (r *Registration) isEmailValid(email string) bool {
	if len(email) < 3 && len(email) > 255 {
		return false
	}

	return emailRegex.MatchString(email)
}
