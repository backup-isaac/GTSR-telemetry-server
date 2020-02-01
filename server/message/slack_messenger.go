package message

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/nlopes/slack"
)

// SlackMessenger handles sending and receiving messages to and from Slack
type SlackMessenger struct {
	client *slack.Client
}

// NewSlackMessenger returns a new SlackMessenger initialized with the provided
// Slack client
func NewSlackMessenger() *SlackMessenger {
	messenger := &SlackMessenger{}
	if token, err := ioutil.ReadFile("/secrets/slack_token.txt"); err == nil {
		messenger.client = slack.New(strings.TrimSpace(string(token)))
	} else {
		log.Printf("Unable to find Slack credentials: %s\n", err)
	}
	return messenger
}

// PostNewMessage posts the provided message to the "chat" channel
func (s *SlackMessenger) PostNewMessage(message string) {
	if s.client != nil {
		s.client.PostMessage("chat", slack.MsgOptionText(message, false))
	} else {
		log.Printf("No slack key - message: %s", message)
	}
}

// RespondToSlackRequest gives feedback for a "chat slash command"
func (s *SlackMessenger) RespondToSlackRequest(text string, res http.ResponseWriter) {
	response := make(map[string]string)
	response["response_type"] = "in_channel"
	response["text"] = text

	res.Header().Set("Content-Type", "application/json")
	json.NewEncoder(res).Encode(response)
}
