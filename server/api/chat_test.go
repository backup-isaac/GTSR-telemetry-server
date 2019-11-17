package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const (
	chatSlashCommandBaseURI = "/chatSlashCommand"
	unauthorizedUser        = "foo"
	authorizedUser          = "U0JSX098T" // Jared
	defaultChannelName      = "chat"
	defaultResponseType     = "in_channel"
)

type chatSlashCommandQueryParams struct {
	channelName string
	userID      string
	text        string
}

type chatSlashCommandResponse struct {
	ResponseType string `json:"response_type"`
	Text         string `json:"text"`
}

func TestChatSlashCommandErrors(t *testing.T) {
	for _, tc := range []struct {
		title            string
		queryParams      chatSlashCommandQueryParams
		expectedResponse chatSlashCommandResponse
	}{{
		title: "Unauthorized user",
		queryParams: chatSlashCommandQueryParams{
			userID:      unauthorizedUser,
			channelName: "foo",
			text:        "bar",
		},
		expectedResponse: chatSlashCommandResponse{
			ResponseType: defaultResponseType,
			Text:         "Unauthorized user",
		},
	}, {
		title: "Must make requests from the #chat channel",
		queryParams: chatSlashCommandQueryParams{
			userID:      authorizedUser,
			channelName: "foo",
			text:        "bar",
		},
		expectedResponse: chatSlashCommandResponse{
			ResponseType: defaultResponseType,
			Text:         "Requests must be made from #chat",
		},
	}, {
		title: "No message provided",
		queryParams: chatSlashCommandQueryParams{
			userID:      authorizedUser,
			channelName: "chat",
			text:        "",
		},
		expectedResponse: chatSlashCommandResponse{
			ResponseType: defaultResponseType,
			Text:         "Please provide a message",
		},
	}, {
		title: "Long message",
		queryParams: chatSlashCommandQueryParams{
			userID:      authorizedUser,
			channelName: "chat",
			text:        "What the fuck did you just fucking say about me, you little bitch? I'll have you know I graduated top of my class in the Navy Seals",
		},
		expectedResponse: chatSlashCommandResponse{
			ResponseType: defaultResponseType,
			Text:         "Message exceeds maximum length",
		},
	}} {
		t.Run(tc.title, func(t *testing.T) {
			// Setting up infrastructure to call the ChatSlashCommand handler
			uri := chatSlashCommandBaseURI +
				"?user_id=" + tc.queryParams.userID +
				"&channel_name=" + tc.queryParams.channelName +
				"&text=" + tc.queryParams.text
			req, err := http.NewRequest("GET", uri, nil)
			if err != nil {
				t.Errorf("Error constructing chatSlashCommand uri. Made: %v", uri)
			}
			rr := httptest.NewRecorder()
			chatHandler := NewChatHandler()
			requireAuthorization = true
			// Call the handler
			chatHandler.ChatSlashCommand(rr, req)
			// Compare status code with expected status code
			resStatusCode := rr.Result().StatusCode
			if resStatusCode != http.StatusOK {
				t.Errorf("Handler responded with non-200 status code: %v", resStatusCode)
			}
			// Sanitize the response body
			resBody := strings.TrimSuffix(rr.Body.String(), "\n")
			// Sanitize/prepare expected response body
			expectedResBody, err := json.Marshal(tc.expectedResponse)
			if err != nil {
				t.Errorf("Failed to marshal expected response body into JSON: %q", err)
			}
			// Compare response body with expected response body
			if strings.Compare(resBody, string(expectedResBody)) != 0 {
				t.Errorf("Unexpected response: want: %q, got: %q", expectedResBody, resBody)
			}
		})
	}
}
