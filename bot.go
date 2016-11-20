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

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
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
		err := http.ListenAndServe(addr, mux)
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
func (b Bot) SendMessage(recipient int, message string, mode ParseMode, preview bool, opts *SendOptions) error {
	urlvalues := url.Values{
		"chat_id":                  {strconv.Itoa(recipient)},
		"text":                     {message},
		"parse_mode":               {string(mode)},
		"disable_web_page_preview": {strconv.FormatBool(!preview)},
	}
	if opts != nil && (opts.ReplyMarkup.Keyboard != nil || opts.ReplyMarkup.ForceReply || opts.ReplyMarkup.Hide) {
		replymarkup, _ := json.Marshal(opts.ReplyMarkup)
		urlvalues.Set("reply_markup", string(replymarkup))
	}

	resp, err := http.PostForm(baseURL+b.token+"/sendMessage", urlvalues)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var r struct {
		OK      bool   `json:"ok"`
		Desc    string `json:"description"`
		ErrCode int    `json:"errorcode"`
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
func (b Bot) SendPhoto(recipient int, photo Photo, caption string, opts *SendOptions) error {
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

	w.WriteField("chat_id", strconv.Itoa(recipient))
	if err := w.Close(); err != nil {
		return err
	}

	resp, err = http.Post(baseURL+b.token+"/sendPhoto", w.FormDataContentType(), &buf)
	if err != nil {
		return fmt.Errorf("Error while sending image to Telegram servers: %v", err)
	}
	defer resp.Body.Close()

	var r struct {
		OK      bool   `json:"ok"`
		ErrCode int    `json:"error_code"`
		Desc    string `json:"description"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return fmt.Errorf("Error while decoding response: %v", err)
	}
	if !r.OK {
		return fmt.Errorf("Error returned from Telegram servers after sending photo: %v (ErrorCode: %v)", r.Desc, r.ErrCode)
	}
	return nil
}

// TODO(ig): implement
//
// SendAudio sends audio files, if you want Telegram clients to display
// them in the music player. audio must be in the .mp3 format and must not
// exceed 50 MB in size.
func (b Bot) sendAudio(recipient User, audio Audio, opts *SendOptions) error {
	panic("not implemented yet")
}

// TODO(ig): implement
//
// SendDocument sends general files. Documents must not exceed 50 MB in size.
func (b Bot) sendDocument(recipient User, document Document, opts *SendOptions) error {
	panic("not implemented yet")
}

// TODO(ig): implement
//
//SendSticker sends stickers with .webp extensions.
func (b Bot) sendSticker(recipient User, sticker Sticker, opts *SendOptions) error {
	panic("not implemented yet")
}

// TODO(ig): implement
//
// SendVideo sends video files. Telegram clients support mp4 videos (other
// formats may be sent as Document). Video files must not exceed 50 MB in size.
func (b Bot) sendVideo(recipient User, video Video, opts *SendOptions) error {
	panic("not implemented yet")
}

// TODO(ig): implement
//
// SendVoice sends audio files, if you want Telegram clients to display
// the file as a playable voice message. For this to work, your audio must be
// in an .ogg file encoded with OPUS (other formats may be sent as Audio or
// Document). audio must not exceed 50 MB in size.
func (b Bot) sendVoice(recipient User, audio Audio, opts *SendOptions) error {
	panic("not implemented yet")
}

// TODO(ig): implement
//
// SendLocation sends location point on the map.
func (b Bot) SendLocation(recipient int, location Location, opts *SendOptions) error {
	urlvalues := url.Values{
		"chat_id":   {strconv.Itoa(recipient)},
		"latitude":  {strconv.FormatFloat(location.Lat, 'f', -1, 64)},
		"longitude": {strconv.FormatFloat(location.Long, 'f', -1, 64)},
	}
	resp, err := http.PostForm(baseURL+b.token+"/sendLocation", urlvalues)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var r struct {
		OK      bool   `json:"ok"`
		Desc    string `json:"description"`
		ErrCode int    `json:"errorcode"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return err
	}
	if !r.OK {
		return fmt.Errorf("%v (%v)", r.Desc, r.ErrCode)
	}
	return nil
}

// SendVenue Use this method to send information about a venue
func (b Bot) SendVenue(recipient int, venue Venue, opts *SendOptions) error {
	urlvalues := url.Values{
		"chat_id":   {strconv.Itoa(recipient)},
		"latitude":  {strconv.FormatFloat(venue.Location.Lat, 'f', -1, 64)},
		"longitude": {strconv.FormatFloat(venue.Location.Long, 'f', -1, 64)},
		"title":     {venue.Title},
		"address":   {venue.Address},
	}
	resp, err := http.PostForm(baseURL+b.token+"/sendVenue", urlvalues)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var r struct {
		OK      bool   `json:"ok"`
		Desc    string `json:"description"`
		ErrCode int    `json:"errorcode"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return err
	}
	if !r.OK {
		return fmt.Errorf("%v (%v)", r.Desc, r.ErrCode)
	}
	return nil
}

// SendChatAction broadcasts type of action to recipient, such as `typing`,
// `uploading a photo` etc.
func (b Bot) SendChatAction(recipient int, action Action) error {
	urlvalues := url.Values{
		"chat_id": {strconv.Itoa(recipient)},
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

	ReplyMarkup ReplyMarkup
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
