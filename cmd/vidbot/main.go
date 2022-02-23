package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/Hargeon/vidbot/pkg/cache"
	"github.com/Hargeon/vidbot/pkg/service/telegram"
	"github.com/Hargeon/vidbot/pkg/service/videocmprs"
	"github.com/go-redis/redis/v8"

	"github.com/joho/godotenv"
)

type RespondedContext struct {
	ctx    context.Context
	done   chan bool
	cancel context.CancelFunc
}

func main() {
	if err := setup(); err != nil {
		panic(err)
	}
}

func setup() error {
	if err := godotenv.Load(); err != nil {
		return err
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_URL"),
		DB:   0,
	})

	_, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		return nil
	}

	rCtx := &RespondedContext{
		done: make(chan bool),
	}

	rCtx.ctx, rCtx.cancel = context.WithCancel(context.Background())
	defer func() {
		// kill telegram local server
		rCtx.cancel()

		// wait untill telegram server is not shutdown
		select {
		case <-rCtx.done:
			fmt.Println("Successfully closed")
		}
	}()

	go runTelegramServer(rCtx.ctx, rCtx.done)

	vSrv := new(videocmprs.Client)

	bot, err := telegram.NewBot(os.Getenv("TELEGRAM_TOKEN"), cache.NewService(rdb), vSrv)
	if err != nil {
		return err
	}

	go bot.ReadMessages()

	fmt.Println("The app is started")
	ext := make(chan os.Signal, 1)
	signal.Notify(ext, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-ext

	return nil
}

func runTelegramServer(ctx context.Context, done chan<- bool) {
	bin := "./bin/mac/telegram-bot-api"
	appID := fmt.Sprintf("--api-id=%s", os.Getenv("TELEGRAM_API_ID"))
	appHash := fmt.Sprintf("--api-hash=%s", os.Getenv("TELEGRAM_API_HASH"))
	l := "--local"

	cmd := exec.CommandContext(ctx, bin, appID, appHash, l)
	err := cmd.Start()
	if err != nil {
		log.Println("in run telegram", err)
	}

	select {
	case <-ctx.Done():
		fmt.Println("Telegram api successfully closed")
		if err = cmd.Process.Kill(); err != nil {
			fmt.Println("err when kill proc", err)
		}
		done <- true
	}
}
