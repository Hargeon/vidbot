package service

import (
	"context"
	"os"
	"time"

	"github.com/Hargeon/vidbot/pkg/service/videocmprs"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot interface {
}

type MessageHandler interface {
	Execute(int64, *videocmprs.Account, tgbotapi.Update)
	SetNext(MessageHandler)
}

type Cache interface {
	Read(key string) (string, error)
	Write(key string, value interface{}, exp time.Duration) error
	Delete(key string) error
}

type VideoService interface {
	Authenticate(chatID int64) (string, error)
	Registration(chatID int64, email string) error
	SendVideo(file *os.File, tokenAuth string, req *videocmprs.Request) error
}

type CloudStorage interface {
	Download(ctx context.Context, id string) (string, error)
}
