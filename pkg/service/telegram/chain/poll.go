package chain

import (
	"fmt"

	"github.com/Hargeon/vidbot/pkg/service/videocmprs"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Poll struct {
	Base
}

// Execute function uses for processing poll
func (p *Poll) Execute(chatID int64, acc *videocmprs.Account, update tgbotapi.Update) {
	if update.PollAnswer == nil {
		p.Next.Execute(chatID, acc, update)

		return
	}

	if acc.State != videocmprs.Attributes {
		p.SomethingWentWrong(chatID)

		return
	}

	opts := update.PollAnswer.OptionIDs
	attrs := make([]string, 0, 5)
	for i := 0; i < len(opts); i++ {
		switch opts[i] {
		case 0: //resolution
			attrs = append(attrs, "resolution_x", "resolution_y")
		case 1: // ratio
			attrs = append(attrs, "ratio_x", "ratio_y")
		case 2: // bitrate
			attrs = append(attrs, "bitrate")
		}
	}

	acc.Attributes = attrs
	fmt.Println("Acc Attribute", acc.Attributes)

	p.Next.Execute(chatID, acc, update)
}
