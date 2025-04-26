package groupme

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

type GroupMeService struct {
	Client *http.Client
}

func NewGroupMeService() *GroupMeService {
	return &GroupMeService{
		Client: &http.Client{
			Timeout: 20 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns: 10,
			},
		},
	}
}

func (g *GroupMeService) SendMessage(message Message, response string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)

	defer cancel()

	url := url.URL{
		Scheme: "https",
		Host:   "api.groupme.com",
		Path:   "/v3/bots/post",
	}

	msg := fmt.Sprintf("@%s %s", message.Name, response)

	messageSendRequest := MessageSendRequest{
		BotId: os.Getenv("GROUPME_BOT_ID"),
		Text:  msg,
	}

	reqBody, err := json.Marshal(messageSendRequest)

	if err != nil {
		return false, fmt.Errorf("failed to marshal request body %v", err)
	}

	params := &http.Request{
		Header: map[string][]string{
			"Content-Type":  {"application/json"},
			"Authorization": {os.Getenv("GROUPME_BOT_TOKEN")},
		},
		Body:   io.NopCloser(bytes.NewReader(reqBody)),
		Method: "POST",
		URL:    &url,
	}

	params = params.WithContext(ctx)

	resp, err := g.Client.Do(params)

	if err != nil {
		return false, fmt.Errorf("failed to send message %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != 202 {
		body, _ := io.ReadAll(resp.Body)
		return false, fmt.Errorf("failed to send message %d, %s", resp.StatusCode, body)
	}

	return true, nil
}
