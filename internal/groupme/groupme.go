package groupme

import (
	"bytes"
	"context"
	"crowfather/internal/config"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type GroupMeService struct {
	Client *http.Client
	Config *config.GroupMeConfig
}

func NewGroupMeService(config *config.GroupMeConfig) *GroupMeService {
	return &GroupMeService{
		Client: &http.Client{
			Timeout: 20 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns: 10,
			},
		},
		Config: config,
	}
}

func (g *GroupMeService) SendMessage(message Message, response string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), g.Config.Timeout)

	defer cancel()

	reqBody, err := g.buildPayload(message, response)

	if err != nil {
		return false, fmt.Errorf("failed to build request body %v", err)
	}

	request := g.buildRequest(ctx, reqBody)

	resp, err := g.sendRequest(request)

	if err != nil || !resp {
		return false, fmt.Errorf("failed to send message %v", err)
	}

	return true, nil
}

func (g *GroupMeService) buildUrl() *url.URL {
	return &url.URL{
		Scheme: "https",
		Host:   g.Config.Host,
		Path:   g.Config.Path,
	}
}

func (g *GroupMeService) buildPayload(message Message, response string) ([]byte, error) {
	return json.Marshal(MessageSendRequest{
		BotId: g.Config.BotID,
		Text:  fmt.Sprintf("@%s %s", message.Name, response),
	})
}

func (g *GroupMeService) buildRequest(ctx context.Context, payload []byte) *http.Request {
	params := &http.Request{
		Header: map[string][]string{
			"Content-Type":  {"application/json"},
			"Authorization": {g.Config.Token},
		},
		Body:   io.NopCloser(bytes.NewReader(payload)),
		Method: "POST",
		URL:    g.buildUrl(),
	}

	params = params.WithContext(ctx)

	return params
}

func (g *GroupMeService) sendRequest(req *http.Request) (bool, error) {
	resp, err := g.Client.Do(req)

	if err != nil {
		return false, fmt.Errorf("failed to send request %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		body, _ := io.ReadAll(resp.Body)
		return false, fmt.Errorf("failed to send request status=%d, body=%s", resp.StatusCode, body)
	}

	return true, nil
}
