package chain

import (
	"github.com/Hargeon/vidbot/pkg/service/videocmprs"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Wrong struct {
	Base
}

// Execute function send error to telegram
func (w *Wrong) Execute(chatID int64, acc *videocmprs.Account, update tgbotapi.Update) {
	w.SomethingWentWrong(chatID)
}
