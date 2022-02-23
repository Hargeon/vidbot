package telegram

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/Hargeon/vidbot/pkg/service"
	"github.com/Hargeon/vidbot/pkg/service/videocmprs"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	reConTimeout       = time.Second * 2
	reconTimes         = 5
	registrationPrefix = "/start "
	expTime            = time.Hour
	baseInt            = 10
	bitSize            = 64
	smthWntWrong       = "Something went wrong. "
	tryOneMore         = "Try one more time. "
	interValidValue    = "Inter valid %s value"
)

var pollQuestions = []string{"Resolution", "Aspect Ratio", "Bitrate"}

type Bot struct {
	bot   *tgbotapi.BotAPI
	cache service.Cache
	// vSrv - client for videocmprs service
	vSrv service.VideoService

	botEnd string
	token  string
}

func NewBot(token string, c service.Cache, vSrv service.VideoService) (*Bot, error) {
	tgHost := os.Getenv("TELEGRAM_HOST")
	botEnd := tgHost + "/bot%s/%s"

	bb := &Bot{cache: c, botEnd: botEnd, token: token, vSrv: vSrv}

	bot, err := reconectBot(token, botEnd)

	bb.bot = bot

	return bb, err
}

func reconectBot(token, botEnd string) (*tgbotapi.BotAPI, error) {
	var tm int
	var bot *tgbotapi.BotAPI
	var err error
	for i := 0; i < reconTimes; i++ {
		bot, err = tgbotapi.NewBotAPIWithAPIEndpoint(token, botEnd)

		if err == nil {
			break
		}

		tm++
		time.Sleep(reConTimeout)
	}

	return bot, err
}

func (b *Bot) ReadMessages() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.bot.GetUpdatesChan(u)

	for u := range updates {
		go b.processMsg(u)
	}
}

func (b *Bot) processMsg(update tgbotapi.Update) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("err recover", err)
			fmt.Println("stacktrace from panic: ", string(debug.Stack()))
		}
	}()

	if update.Message != nil {
		switch {
		case update.Message.Text == "/start":
			b.comandStart(update)
		case update.Message.Video != nil:
			b.processVideo(update)
		default:
			if update.Message != nil && strings.HasPrefix(update.Message.Text, registrationPrefix) {
				// registration
				b.startWithToken(update)
				return
			}

			chatID := update.Message.Chat.ID
			acc, err := b.readFromCache(chatID)
			if err != nil {
				b.authAndInitState(chatID)
				return
			}

			b.chooseAction(chatID, acc, update)
		}
	} else if update.PollAnswer != nil {
		b.pollAnswer(update)
	}
}

func (b *Bot) chooseAction(chatID int64, acc *videocmprs.Account, update tgbotapi.Update) {
	if acc.State == videocmprs.Attributes {
		if acc.Request == nil {
			acc.Request = new(videocmprs.Request)
		}

		switch acc.CurrentAttr {
		case "resolution_x":
			resX, err := strconv.Atoi(update.Message.Text)
			if err != nil {
				fmt.Println("resolution_x parse error", err)
				msg := tgbotapi.NewMessage(chatID, smthWntWrong+fmt.Sprintf(interValidValue, acc.CurrentAttr))
				b.sendMsg(msg)

				return
			}
			acc.Request.ResolutionX = resX
		case "resolution_y":
			resY, err := strconv.Atoi(update.Message.Text)
			if err != nil {
				fmt.Println("resolution_y parse error", err)
				msg := tgbotapi.NewMessage(chatID, smthWntWrong+fmt.Sprintf(interValidValue, acc.CurrentAttr))
				b.sendMsg(msg)

				return
			}
			acc.Request.ResolutionY = resY
		case "ratio_x":
			ratX, err := strconv.Atoi(update.Message.Text)
			if err != nil {
				fmt.Println("ratio_x parse error", err)
				msg := tgbotapi.NewMessage(chatID, smthWntWrong+fmt.Sprintf(interValidValue, acc.CurrentAttr))
				b.sendMsg(msg)

				return
			}
			acc.Request.RatioX = ratX
		case "ratio_y":
			ratY, err := strconv.Atoi(update.Message.Text)
			if err != nil {
				fmt.Println("ratio_y parse error", err)
				msg := tgbotapi.NewMessage(chatID, smthWntWrong+fmt.Sprintf(interValidValue, acc.CurrentAttr))
				b.sendMsg(msg)

				return
			}
			acc.Request.RatioY = ratY
		case "bitrate":
			bit, err := strconv.ParseInt(update.Message.Text, baseInt, bitSize)
			if err != nil {
				fmt.Println("bitrate parse error", err)
				msg := tgbotapi.NewMessage(chatID, smthWntWrong+fmt.Sprintf(interValidValue, acc.CurrentAttr))
				b.sendMsg(msg)

				return
			}
			acc.Request.Bitrate = bit
		}

		b.addAtribute(chatID, acc)
	} else if acc.State == videocmprs.Video {
		b.sendVideo(chatID)
	}
}

func (b *Bot) pollAnswer(update tgbotapi.Update) {
	id := update.PollAnswer.User.ID
	fmt.Println("poll answ", id)
	acc, err := b.readFromCache(id)
	if err != nil {
		// try auth

		// add to cache with attr
		fmt.Println("Err in read cache", err)
		return
	}
	if acc.State != videocmprs.Attributes {
		// choose right step
		fmt.Println("acc step is ", acc.State)
	}
	opts := update.PollAnswer.OptionIDs
	fmt.Println("poll opts", opts)
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

	b.addAtribute(id, acc)
}

func (b *Bot) addAtribute(chatID int64, acc *videocmprs.Account) {
	if acc.Attributes == nil {
		acc.State = videocmprs.Video
		if err := b.addToCache(acc); err != nil {
			b.somethingWentWrong(chatID)
		}

		b.sendVideo(chatID)
		fmt.Println("Send video")
		return
	}

	var attr string
	if len(acc.Attributes) == 1 {
		attr = acc.Attributes[0]
		acc.Attributes = nil
		// send video
	} else {
		attr = acc.Attributes[0]
		acc.Attributes = acc.Attributes[1:]
	}

	acc.CurrentAttr = attr
	if err := b.addToCache(acc); err != nil {
		b.somethingWentWrong(chatID)
		return
	}

	msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Add %s value", attr))
	b.bot.Send(msg)
}

func (b *Bot) sendVideo(chatID int64) {
	acc, err := b.readFromCache(chatID)
	if err != nil {
		// something went wrong
		b.somethingWentWrong(chatID)
	}
	if acc.State != videocmprs.Video {
		acc.State = videocmprs.Video
		err = b.addToCache(acc)
		if err != nil {
			b.somethingWentWrong(chatID)
		}
	}
	msg := tgbotapi.NewMessage(chatID, "Please, send video file")
	b.bot.Send(msg)
}

func (b *Bot) comandStart(update tgbotapi.Update) {
	fmt.Println("update.Message.Chat.ID", update.Message.Chat.ID)
	fmt.Println("in command")
	log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

	poll := tgbotapi.NewPoll(update.Message.Chat.ID, "What do you want to change",
		"Resolution", "Aspect Ratio", "Bitrate")
	poll.AllowsMultipleAnswers = true
	poll.IsAnonymous = false

	b.bot.Send(poll)
}

func (b *Bot) sendPoll(chatID int64) {
	poll := tgbotapi.NewPoll(chatID, "What do you want to change", pollQuestions...)
	poll.AllowsMultipleAnswers = true
	poll.IsAnonymous = false

	b.bot.Send(poll)
}

func (b *Bot) processVideo(update tgbotapi.Update) {
	id := update.Message.Video.FileID
	chatID := update.Message.Chat.ID
	acc, err := b.readFromCache(chatID)
	if err != nil {
		b.somethingWentWrong(chatID)

		return
	}

	if acc.State != videocmprs.Video {
		b.somethingWentWrong(chatID)
		b.authAndInitState(chatID)

		return
	}
	fmt.Println("update.Message.Video", update.Message.Video)

	file, err := b.getFile(id)
	if err != nil {
		fmt.Println("err in get file", err)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "We got an error when recieving your file")
		b.bot.Send(msg)
		return
	}

	fmt.Println("File", file)

	if err = b.vSrv.SendVideo(file, acc.TokenAuth, acc.Request); err != nil {
		fmt.Println("Err in send video", err)
		b.somethingWentWrong(chatID)
		return
	}
	b.deleteFromCache(chatID)

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "We got your video")
	b.bot.Send(msg)
}

func (b *Bot) getFile(fileID string) (*os.File, error) {
	file, err := b.bot.GetFile(tgbotapi.FileConfig{FileID: fileID})

	if err != nil {
		return nil, err
	}

	fmt.Println("file path", file.FilePath)

	return os.Open(file.FilePath)
}

func (b *Bot) startWithToken(update tgbotapi.Update) {
	token := strings.TrimPrefix(update.Message.Text, registrationPrefix)
	chatID := update.Message.Chat.ID

	account, err := b.readFromCache(chatID)
	if err != nil {
		// try auth
		tokenAuth, err := b.vSrv.Authenticate(chatID)
		if err != nil {
			fmt.Println("err in authenticate", err)
			if errors.Is(err, videocmprs.ErrNotRegistered) {
				// try register
				err = b.vSrv.Registration(chatID, token)
				if err != nil {
					fmt.Println("err after registration", err)
					b.somethingWentWrong(chatID)
					return
				}

				b.authAndInitState(chatID)
				return
			}

			b.somethingWentWrong(chatID)
			return
		}

		b.initState(chatID, tokenAuth)

		return
	}

	// check step and do action

	fmt.Println("finish", account)

	msg := tgbotapi.NewMessage(chatID, "Telegram account successfylly added")
	b.sendMsg(msg)
}

func (b *Bot) initState(chatID int64, tokenAuth string) {
	account := &videocmprs.Account{
		TokenAuth: tokenAuth,
		State:     videocmprs.Attributes,
		ChatID:    chatID,
	}

	err := b.addToCache(account)
	if err != nil {
		b.somethingWentWrong(chatID)
	}

	b.sendPoll(chatID)
}

func (b *Bot) deleteFromCache(chatID int64) error {
	key := fmt.Sprintf("%d", chatID)

	return b.cache.Delete(key)
}

func (b *Bot) authAndInitState(chatID int64) {
	tokenAuth, err := b.vSrv.Authenticate(chatID)
	if err != nil {
		fmt.Println("err in second auth", err)
		b.somethingWentWrong(chatID)
		return
	}

	b.initState(chatID, tokenAuth)
}

func (b *Bot) sendMsg(c tgbotapi.Chattable) {
	b.bot.Send(c)
}

func (b *Bot) readFromCache(chatID int64) (*videocmprs.Account, error) {
	key := fmt.Sprintf("%d", chatID)
	resStr, err := b.cache.Read(key)
	if err != nil {
		return nil, err
	}
	fmt.Println("read cache key", key, "value", resStr)

	res := new(videocmprs.Account)
	err = json.Unmarshal([]byte(resStr), res)

	return res, err
}

func (b *Bot) addToCache(acc *videocmprs.Account) error {
	res, err := json.Marshal(acc)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("%d", acc.ChatID)
	fmt.Println("add to cache key", key, "value", string(res))
	return b.cache.Write(key, res, expTime)
}

func (b *Bot) somethingWentWrong(chatID int64) {
	msg := tgbotapi.NewMessage(chatID, "Somening went wrong, try later")
	b.sendMsg(msg)
}
