package service

import (
	"os"
	"time"

	"github.com/Hargeon/vidbot/pkg/service/videocmprs"
)

type Bot interface {
}

type Cache interface {
	Read(key string) (string, error)
	Write(key string, value interface{}, exp time.Duration) error
	Delete(key string) error
}

type VideoService interface {
	Authenticate(chatID int64) (string, error)
	Registration(chatID int64, token string) error
	SendVideo(file *os.File, tokenAuth string, req *videocmprs.Request) error
}
