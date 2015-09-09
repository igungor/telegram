package tlbot

import (
	"bytes"
	"fmt"
)

type Action string

// Types of actions to broadcast
const (
	Typing            Action = "typing"
	UploadingPhoto    Action = "upload_photo"
	UploadingVideo    Action = "upload_video"
	UploadingAudio    Action = "upload_audio"
	UploadingDocument Action = "upload_document"
	FindingLocation   Action = "find_location"
)

// User represents a Telegram user or bot.
type User struct {
	ID        int    `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`

	Title string `json:"title"`
}

// IsGroupChat reports whether the message is originally sent from a chat group.
//
// Telegram can send User or GroupChat interchangebly depending on the
// original sender. If title is not empty, it means the message is
// originally sent from a chat group.
func (u User) IsGroupChat() bool { return u.Title != "" }

type Update struct {
	ID      int     `json:"update_id"`
	Payload Message `json:"message"`
}

// Message represents a message to be sent.
type Message struct {
	// Unique message identifier
	ID int `json:"message_id"`

	// Sender
	From User `json:"from"`

	// Date is when the message was sent in Unix time
	Date int `json:"date"`

	// Conversation the message belongs to — user in case of a private message,
	// GroupChat in case of a group
	Chat User `json:"chat"`

	// For forwarded messages, sender of the original message (Optional)
	ForwardFrom User `json:"forward_from"`

	// For forwarded messages, date the original message was sent in
	// Unix time (Optional)
	ForwardDate int `json:"forward_date"`

	// For replies, the original message. Note that the Message
	// object in this field will not contain further reply_to_message fields
	// even if it itself is a reply (Optional)
	ReplyTo *Message `json:"reply_to_message"`

	// For text messages, the actual UTF-8 text of the message (Optional)
	Text string `json:"text"`

	// Message is an audio file, information about the file (Optional)
	Audio Audio `json:"audio"`

	// Message is a general file, information about the file (Optional)
	Document Document `json:"document"`

	// Message is a photo, available sizes of the photo (Optional)
	Photos []Photo `json:"photo"`

	// Message is a sticker, information about the sticker (Optional)
	Sticker Sticker `json:"sticker"`

	// Message is a video, information about the video (Optional)
	Video Video `json:"video"`

	// Message is a shared contact, information about the contact (Optional)
	Contact Contact `json:"contact"`

	// Message is a shared location, information about the location (Optional)
	Location Location `json:"location"`

	// A new member was added to the group, information about them
	// (this member may be bot itself) (Optional)
	NewChatParticipant User `json:"new_chat_participant"`

	// A member was removed from the group, information about them
	// (this member may be bot itself) (Optional)
	LeftChatParticipant User `json:"left_chat_participant"`

	// A group title was changed to this value (Optional)
	NewChatTitle string `json:"new_chat_title"`

	// A group photo was change to this value (Optional)
	NewChatPhoto []Photo `json:"new_chat_photo"`

	// Informs that the group photo was deleted (Optional)
	DeleteChatPhoto bool `json:"delete_chat_photo"`

	// Informs that the group has been created (Optional)
	GroupChatCreated bool `json:"group_chat_created"`
}

func (m Message) String() string {
	var buf bytes.Buffer
	if m.From.IsGroupChat() {
		buf.WriteString(fmt.Sprintf("From group: %q  ", m.From.Title))
	} else {
		buf.WriteString(fmt.Sprintf(`From user: "%v %v (%v)"  `, m.From.FirstName, m.From.LastName, m.From.Username))
	}
	buf.WriteString(fmt.Sprintf("Message: %q\n", m.Text))

	return buf.String()
}

type File struct {
	// File is embedded in most of the types. So a `File` prefix is used
	FileID   string `json:"file_id"`
	FileSize int    `json:"file_size"`

	// URL to the file (custom field)
	FileURL string `json:"-"`

	// Path to the file on local filesystem (custom field)
	FilePath string `json:"-"`
}

// Exists reports whether the file is already at Telegram servers.
func (f File) Exists() bool { return f.FileID != "" }

// IsLocal reports whether the file is the local filesystem.
func (f File) IsLocal() bool { return f.FilePath != "" }

// IsRemote reports whether the file is on a remote server.
func (f File) IsRemote() bool { return f.FileURL != "" }

// Photo represents one size of a photo or a file/sticker thumbnail.
type Photo struct {
	File
	Width  int `json:"width"`
	Height int `json:"height"`
}

// Audio represents an audio file.
type Audio struct {
	File
	Duration  int    `json:"duration"`
	Performer string `json:"performer"`
	Title     string `json:"title"`
	MimeType  string `json:"mime_type"`
}

// Document represents a general file (as opposed to photos and audio files).
type Document struct {
	File
	Filename  string `json:"file_name"`
	Thumbnail Photo  `json:"thumb"`
	MimeType  string `json:"mime_type"`
}

// Sticker represents a sticker.
type Sticker struct {
	File
	Width     int   `json:"width"`
	Height    int   `json:"height"`
	Thumbnail Photo `json:"thumb"`
}

// Video represents a video file.
type Video struct {
	File
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	Duration  int    `json:"duration"`
	Thumbnail Photo  `json:"thumb"`
	MimeType  string `json:"mime_type"`
	Caption   string `json:"caption"`
}

// Voice represents an voice note.
type Voice struct {
	File
	Duration int    `json:"duration"`
	MimeType string `json:"mime_type"`
}

// Location represents a point on the map.
type Location struct {
	Lat  float32 `json:"latitude"`
	Long float32 `json:"longitude"`
}

// Contact represents a phone contact.
type Contact struct {
	PhoneNumber string `json:"phone_number"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	UserID      string `json:"user_id"`
}

type ReplyKeyboardMarkup struct {
	Keyboard  [][]string `json:"keyboard"`
	Resize    bool       `json:"resize_keyboard"`
	OneTime   bool       `json:"one_time_keyboard"`
	Selective bool       `json:"selective"`
}

// ReplyKeyboardHide requests Telegram clients to hide the current custom
// keyboard and display the default letter-keyboard
type ReplyKeyboardHide struct {
	Hide      bool `json:"hide_keyboard"`
	Selective bool `json:"selective"`
}

// ForceReply forces Telegram clients to display a reply interface to the user
// (act as if the user has selected the bot‘s message and tapped ’Reply').
type ForceReply struct {
	ForceReply bool `json:"force_reply"`
	Selective  bool `json:"selective"`
}
