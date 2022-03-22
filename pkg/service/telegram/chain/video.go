package chain

import (
	"fmt"
	"os"

	"github.com/Hargeon/vidbot/pkg/service/videocmprs"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	videoErrMsg    = "We got an error when recieving your file"
	gotVideoMsg    = "We got your video. Processing..."
	sendVideoMsg   = "Please, send a video"
	receivingVIdeo = "Receiving your video..."
)

type Video struct {
	Base
}

// Execute function uses for sending video to videocmprs service
func (v *Video) Execute(chatID int64, acc *videocmprs.Account, update tgbotapi.Update) {
	if acc.State != videocmprs.Video {
		v.Next.Execute(chatID, acc, update)

		return
	}

	if update.Message.Video == nil {
		msg := tgbotapi.NewMessage(chatID, sendVideoMsg)
		v.SendMsg(msg)

		return
	}

	msg := tgbotapi.NewMessage(chatID, receivingVIdeo)
	v.SendMsg(msg)

	fileID := update.Message.Video.FileID
	fmt.Println("file id", fileID)
	file, err := v.getFile(fileID)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, videoErrMsg)
		v.SendMsg(msg)

		return
	}

	if err = v.VSrv.SendVideo(file, acc.TokenAuth, acc.Request); err != nil {
		v.SomethingWentWrong(chatID)

		return
	}

	v.DeleteFromCache(chatID)

	msg = tgbotapi.NewMessage(chatID, gotVideoMsg)
	v.SendMsg(msg)
}

func (v *Video) getFile(fileID string) (*os.File, error) {
	file, err := v.Bot.GetFile(tgbotapi.FileConfig{FileID: fileID})

	if err != nil {
		return nil, err
	}

	return os.Open(file.FilePath)
}
