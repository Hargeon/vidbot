// Package consumer uses for getting response from rabbit
package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/Hargeon/vidbot/pkg/service/storage"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/streadway/amqp"
)

const (
	cantDownloadVideo = "Can't download video from cloud. Please, send message to support."
)

type Consumer struct {
	BotClient *tgbotapi.BotAPI
	Store     storage.AWSS3

	Msgs <-chan amqp.Delivery
}

// HandleMessages function uses for processing messages from rabbit
func (c *Consumer) HandleMessages() {
	for d := range c.Msgs {
		log.Println("Recieve from main service", string(d.Body))

		resp := new(Response)

		if err := json.Unmarshal(d.Body, resp); err != nil {
			log.Println("Err in marshaling body", err)

			continue
		}

		if resp.ChatID == 0 {
			log.Println("Invalid ChatID", resp.ChatID)

			continue
		}

		if resp.Err != "" {
			msg := tgbotapi.NewMessage(resp.ChatID, resp.Err)
			c.sendMsg(msg)

			return
		}

		filepath, err := c.Store.Download(context.Background(), resp.Video.ServiceID)
		if err != nil {
			fmt.Println("download err", err)
			msg := tgbotapi.NewMessage(resp.ChatID, cantDownloadVideo)
			c.sendMsg(msg)
		} else {
			c.sendVideo(resp.ChatID, filepath)
		}

		msgStr := c.buildVideoMsg(resp)
		msg := tgbotapi.NewMessage(resp.ChatID, msgStr)
		msg.ParseMode = tgbotapi.ModeMarkdownV2
		c.sendMsg(msg)

		fmt.Println("Finish msg")
		c.finishMsg(d)
	}
}

func (c *Consumer) buildVideoMsg(resp *Response) string {
	var result string
	if resp.Video.Size != 0 {
		result += fmt.Sprintf("*Size:* %d bytes\n", resp.Video.Size)
	}

	if resp.Video.Bitrate != 0 {
		result += fmt.Sprintf("*Bitrate:* %d k\n", resp.Video.Bitrate)
	}

	if resp.Video.ResolutionX != 0 || resp.Video.ResolutionY != 0 {
		result += fmt.Sprintf("*Resolution:* %d:%d\n", resp.Video.ResolutionX, resp.Video.ResolutionY)
	}

	if resp.Video.RatioX != 0 || resp.Video.RatioY != 0 {
		result += fmt.Sprintf("*Aspect ratio:* %d:%d\n", resp.Video.RatioX, resp.Video.RatioY)
	}

	return result
}

func (c *Consumer) sendMsg(msg tgbotapi.Chattable) {
	if _, err := c.BotClient.Send(msg); err != nil {
		fmt.Println("err when sending msg", err)
	}
}

func (c *Consumer) finishMsg(d amqp.Delivery) {
	if err := d.Ack(false); err != nil { // needs to mark a message was processed
		log.Println("Ack err", err)
	}
}

func (c *Consumer) sendVideo(chatID int64, filepath string) {
	upl := &Uploader{Path: filepath}
	msg := tgbotapi.NewVideo(chatID, upl)
	c.sendMsg(msg)
}
