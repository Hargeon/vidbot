package videocmprs

import (
	"bytes"
	"crypto/sha1"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/jsonapi"
)

const (
	timeOut             = time.Second * 2
	contentType         = "application/vnd.api+json"
	tokenPrefix         = "Bearer"
	multipartType       = "multipart/form-data"
	telegramHeader      = "Telegram-Auth"
	successStatusCode   = 201
	videoRequestTimeOut = time.Minute * 15
)

// Client for videocmprs service
type Client struct{}

// Authenticate telegram account
func (c *Client) Authenticate(chatID int64) (string, error) {
	client := &http.Client{
		Timeout: timeOut,
	}
	url := os.Getenv("SERVICE_URL") + "/auth/tg_account"
	acc := &Account{
		ChatID: chatID,
	}

	buf := new(bytes.Buffer)

	if err := jsonapi.MarshalPayload(buf, acc); err != nil {
		return "", err
	}

	token := c.generateTgToken(buf.String())
	req, err := http.NewRequest(http.MethodPost, url, buf)
	if err != nil {
		return "", err
	}

	req.Header.Add("Content-Type", contentType)
	req.Header.Add("Accept", contentType)
	req.Header.Add(telegramHeader, token)

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusForbidden:
		return "", ErrNotRegistered
	case http.StatusCreated:
		acc = new(Account)
		if err := jsonapi.UnmarshalPayload(resp.Body, acc); err != nil {
			return "", err
		}

		return acc.TokenAuth, nil
	default:
		return "", ErrSomethingWentWrong
	}
}

// Registration telegram account
func (c *Client) Registration(chatID int64, token string) error {
	tg := &Account{
		ChatID: chatID,
		Token:  token,
	}

	buf := new(bytes.Buffer)

	if err := jsonapi.MarshalPayload(buf, tg); err != nil {
		return err
	}

	client := http.Client{
		Timeout: timeOut,
	}

	url := os.Getenv("SERVICE_URL") + "/tg_accounts"

	req, err := http.NewRequest(http.MethodPost, url, buf)
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", contentType)
	req.Header.Add("Accept", contentType)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != successStatusCode {
		return errors.New("can't add telegram account")
	}

	return nil
}

// SendVideo to videocmprs service
func (c *Client) SendVideo(file *os.File, tokenAuth string, req *Request) error {
	bodyBuf := new(bytes.Buffer)
	w := multipart.NewWriter(bodyBuf)

	fw, err := w.CreateFormField("requests")
	if err != nil {
		return err
	}

	reqBuf := new(bytes.Buffer)
	if err := jsonapi.MarshalPayload(reqBuf, req); err != nil {
		return err
	}

	_, err = io.Copy(fw, reqBuf)
	if err != nil {
		return err
	}

	fw, err = c.createFormFile(w, "video", file.Name())
	if err != nil {
		return err
	}

	_, err = io.Copy(fw, file)
	if err != nil {
		return err
	}

	w.Close()

	client := &http.Client{
		Timeout: videoRequestTimeOut,
	}

	url := os.Getenv("SERVICE_URL") + "/requests"
	fmt.Println("URL", url)
	fmt.Println("req body", bodyBuf.Len())
	request, err := http.NewRequest(http.MethodPost, url, bodyBuf)
	if err != nil {
		return err
	}
	request.Header.Add("Content-Type", w.FormDataContentType())
	request.Header.Add("Accept", contentType)
	request.Header.Add("Authorization", fmt.Sprintf("%s %s", tokenPrefix, tokenAuth))

	resp, err := client.Do(request)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Println("resp.StatusCode", resp.StatusCode)
	if resp.StatusCode != http.StatusCreated {
		return ErrCantSendVideo
	}

	return nil
}

// createFormFile is a convenience wrapper around CreatePart. It creates
// a new form-data header with the provided field name and file name.
func (c *Client) createFormFile(w *multipart.Writer, fieldname, filename string) (io.Writer, error) {
	quoteEscaper := strings.NewReplacer("\\", "\\\\", `"`, "\\\"")

	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition",
		fmt.Sprintf(`form-data; name="%s"; filename="%s"`,
			quoteEscaper.Replace(fieldname), quoteEscaper.Replace(filename)))
	h.Set("Content-Type", "application/octet-stream")

	ext := filepath.Ext(filename)
	h.Add("Content-Type", mime.TypeByExtension(ext))

	return w.CreatePart(h)
}

// generateTgToken function generate token auth in videocmprs service
func (c *Client) generateTgToken(body string) string {
	h := sha1.New()
	str := os.Getenv("TELEGRAM_SALT") + body
	h.Write([]byte(str))

	bs := h.Sum(nil)

	return fmt.Sprintf("%x", bs)
}
