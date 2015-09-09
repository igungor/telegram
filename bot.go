package tlbot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

const baseURL = "https://api.telegram.org/bot"

type ParseMode string

// Parse modes
const (
	ModeNone     ParseMode = ""
	ModeMarkdown ParseMode = "markdown"
)

// Bot represent a Telegram bot.
type Bot struct {
	token string
	Info  User
}

// New creates a new Telegram bot with the given token, which is given by
// Botfather. See https://core.telegram.org/bots#botfather
func New(token string) Bot {
	u, _ := getMe(token)

	return Bot{token: token, Info: u}
}

// Listen listens on the given address addr and returns a read-only Message
// channel.
func (b Bot) Listen(addr string) <-chan Message {
	messageCh := make(chan Message)

	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		defer w.WriteHeader(http.StatusOK)

		var u Update
		if err := json.NewDecoder(req.Body).Decode(&u); err != nil {
			log.Printf("error decoding request body: %v\n", err)
			return

		}
		messageCh <- u.Payload
	})

	go func() {
		// ListenAndServe always returns non-nil error
		err := http.ListenAndServe(addr, nil)
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}()

	return messageCh
}

// SetWebhook assigns bot's webhook url with the given url.
func (b Bot) SetWebhook(webhook string) error {
	urlvalues := url.Values{"url": {webhook}}

	resp, err := http.PostForm(baseURL+b.token+"/setWebhook", urlvalues)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var r struct {
		OK      bool   `json:"ok"`
		ErrCode int    `json:"errorcode"`
		Desc    string `json:"description"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return err
	}

	if !r.OK {
		return fmt.Errorf("%v (%v)", r.Desc, r.ErrCode)
	}

	return nil
}

// SendMessage sends text message to the recipient. Callers can send plain
// text or markdown messages by setting mode parameter.
func (b Bot) SendMessage(recipient User, message string, mode ParseMode, preview bool, opts *SendOptions) error {
	urlvalues := url.Values{
		"chat_id":                  {strconv.Itoa(recipient.ID)},
		"text":                     {message},
		"parse_mode":               {string(mode)},
		"disable_web_page_preview": {strconv.FormatBool(!preview)},
	}

	resp, err := http.PostForm(baseURL+b.token+"/sendMessage", urlvalues)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var r struct {
		OK      bool   `json:"ok"`
		ErrCode int    `json:"errorcode"`
		Desc    string `json:"description"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return err
	}

	if !r.OK {
		return fmt.Errorf("%v (%v)", r.Desc, r.ErrCode)
	}

	return nil
}

// TODO(ig): implement
func (b Bot) forwardMessage(recipient User, message Message) error {
	panic("not implemented yet")
}

// SendPhoto sends given photo to recipient. Only remote URLs are supported for now.
// A trivial example is:
//
//  b := bot.New("your-token-here")
//  photo := bot.Photo{FileURL: "http://i.imgur.com/6S9naG6.png"}
//  err := b.SendPhoto(recipient, photo, "sample image", nil)
//
func (b Bot) SendPhoto(recipient User, photo Photo, caption string, opts *SendOptions) error {
	// TODO(ig): implement sending already sent photos
	if photo.Exists() {
		panic("files reside in telegram servers can not be sent for now.")
	}

	// TODO(ig): implement local file upload
	if photo.IsLocal() {
		panic("local files can not be sent for now.")
	}

	resp, err := http.Get(photo.FileURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Fetch failed (errcode: %v). Remote URL: '%v'", resp.StatusCode, photo.FileURL)
	}

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	part, err := w.CreateFormFile("photo", "image.jpg")
	if err != nil {
		return err
	}

	if _, err := io.Copy(part, resp.Body); err != nil {
		return err
	}

	w.WriteField("chat_id", strconv.Itoa(recipient.ID))

	if opts != nil {
		switch opts.ReplyMarkup.(type) {
		case ReplyKeyboardMarkup:
			b, err := json.Marshal(opts.ReplyMarkup)
			if err != nil {
				log.Printf("error while encoding keyboard: %v\n", err)
				return err
			}
			w.WriteField("reply_markup", string(b))

		case ReplyKeyboardHide:
			b, err := json.Marshal(opts.ReplyMarkup)
			if err != nil {
				log.Printf("error while encoding keyboard: %v\n", err)
				return err
			}
			w.WriteField("reply_markup", string(b))

		case ForceReply:
			b, err := json.Marshal(opts.ReplyMarkup)
			if err != nil {
				log.Printf("error while encoding keyboard: %v\n", err)
				return err
			}
			w.WriteField("reply_markup", string(b))

		default:
		}
	}

	if err := w.Close(); err != nil {
		return err
	}

	resp, err = http.Post(baseURL+b.token+"/sendPhoto", w.FormDataContentType(), &buf)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var r struct {
		OK      bool   `json:"ok"`
		ErrCode int    `json:"errorcode"`
		Desc    string `json:"description"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return err
	}

	if !r.OK {
		return fmt.Errorf("%v (%v)", r.Desc, r.ErrCode)
	}

	return nil
}

// TODO(ig): implement
func (b Bot) sendAudio(recipient User, audio Audio, opts *SendOptions) error {
	panic("not implemented yet")
}

// TODO(ig): implement
func (b Bot) sendDocument(recipient User, document Document, opts *SendOptions) error {
	panic("not implemented yet")
}

// TODO(ig): implement
func (b Bot) sendSticker(recipient User, sticker Sticker, opts *SendOptions) error {
	panic("not implemented yet")
}

// TODO(ig): implement
func (b Bot) sendVideo(recipient User, video Video, opts *SendOptions) error {
	panic("not implemented yet")
}

// TODO(ig): implement
func (b Bot) sendVoice(recipient User, audio Audio, opts *SendOptions) error {
	panic("not implemented yet")
}

// TODO(ig): implement
func (b Bot) sendLocation(recipient User, location Location, opts *SendOptions) error {
	panic("not implemented yet")
}

// SendChatAction broadcasts type of action to recipient, such as `typing`,
// `uploading a photo` etc.
func (b Bot) SendChatAction(recipient User, action Action) error {
	urlvalues := url.Values{
		"chat_id": {strconv.Itoa(recipient.ID)},
		"action":  {string(action)},
	}

	resp, err := http.PostForm(baseURL+b.token+"/sendChatAction", urlvalues)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var r struct {
		OK      bool   `json:"ok"`
		ErrCode int    `json:"error_code"`
		Desc    string `json:"description"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil
	}

	if !r.OK {
		return fmt.Errorf("%v (%v)", r.Desc, r.ErrCode)
	}

	return nil
}

type SendOptions struct {
	// If the message is a reply, ID of the original message
	ReplyToMessageID int

	// ReplyKeyboardMarkup || ReplyKeyboardHide || ForceReply
	ReplyMarkup interface{}
}

func getMe(token string) (User, error) {
	resp, err := http.PostForm(baseURL+token+"/getMe", url.Values{})
	if err != nil {
		return User{}, err
	}
	defer resp.Body.Close()

	var r struct {
		OK      bool   `json:"ok"`
		User    User   `json:"result"`
		Desc    string `json:"description"`
		ErrCode int    `json:"error_code"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return User{}, err
	}

	if !r.OK {
		return User{}, fmt.Errorf("%v (%v)", r.Desc, r.ErrCode)
	}

	return r.User, nil
}
