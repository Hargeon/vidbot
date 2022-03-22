package chain

import (
	"fmt"
	"strconv"

	"github.com/Hargeon/vidbot/pkg/service/videocmprs"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	interValidValue = "Inter valid %s value"
	smthWntWrong    = "Something went wrong. "
	baseInt         = 10
	bitSize         = 64
)

type Attribute struct {
	Base
}

// Execute function uses for setting video params
func (a *Attribute) Execute(chatID int64, acc *videocmprs.Account, update tgbotapi.Update) {
	if acc.State != videocmprs.Attributes {
		a.Next.Execute(chatID, acc, update)

		return
	}

	if err := a.setAttribute(acc, update); err != nil {
		msg := tgbotapi.NewMessage(chatID, smthWntWrong+fmt.Sprintf(interValidValue, acc.CurrentAttr))
		a.SendMsg(msg)

		return
	}

	if acc.Attributes == nil {
		acc.State = videocmprs.Video
		if err := a.AddToCache(acc); err != nil {
			a.SomethingWentWrong(chatID)

			return
		}

		msg := tgbotapi.NewMessage(chatID, "Please, send video file")

		a.SendMsg(msg)

		return
	}

	var attr string
	if len(acc.Attributes) == 1 {
		attr = acc.Attributes[0]
		acc.Attributes = nil
		acc.CurrentAttr = ""
	} else {
		attr = acc.Attributes[0]
		acc.Attributes = acc.Attributes[1:]
	}

	acc.CurrentAttr = attr
	if err := a.AddToCache(acc); err != nil {
		a.SomethingWentWrong(chatID)

		return
	}

	msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Add %s value", attr))
	a.SendMsg(msg)
}

func (a *Attribute) setAttribute(acc *videocmprs.Account, update tgbotapi.Update) error {
	if acc.Request == nil {
		acc.Request = new(videocmprs.Request)
	}

	fmt.Println("set attr", acc.CurrentAttr)

	switch acc.CurrentAttr {
	case "resolution_x":
		resX, err := strconv.Atoi(update.Message.Text)
		if err != nil {
			return err
		}

		acc.Request.ResolutionX = resX
	case "resolution_y":
		resY, err := strconv.Atoi(update.Message.Text)
		if err != nil {
			return err
		}

		acc.Request.ResolutionY = resY
	case "ratio_x":
		ratX, err := strconv.Atoi(update.Message.Text)
		if err != nil {
			return err
		}

		acc.Request.RatioX = ratX
	case "ratio_y":
		ratY, err := strconv.Atoi(update.Message.Text)
		if err != nil {
			return err
		}

		acc.Request.RatioY = ratY
	case "bitrate":
		bit, err := strconv.ParseInt(update.Message.Text, baseInt, bitSize)
		if err != nil {
			return err
		}

		acc.Request.Bitrate = bit
	}

	return nil
}
