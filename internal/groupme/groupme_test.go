package groupme

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"crowfather/internal/config"
)

func newService() *GroupMeService {
	return &GroupMeService{Config: &config.GroupMeConfig{BotID: "bot", Token: "token", Host: "example.com", Path: "/v3/bots/post"}, Client: http.DefaultClient}
}

func TestBuildURL(t *testing.T) {
	svc := newService()
	u := svc.buildUrl()
	if u.Scheme != "https" || u.Host != "example.com" || u.Path != "/v3/bots/post" {
		t.Errorf("unexpected url: %v", u)
	}
}

func TestBuildPayload(t *testing.T) {
	svc := newService()
	payload, err := svc.buildPayload(Message{Name: "Bob"}, "hi")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var req MessageSendRequest
	if err := json.Unmarshal(payload, &req); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if req.BotId != "bot" || req.Text != "@Bob hi" {
		t.Errorf("unexpected payload: %+v", req)
	}
}

func TestBuildRequest(t *testing.T) {
	svc := newService()
	req := svc.buildRequest(context.Background(), []byte("data"))
	if req.Method != "POST" {
		t.Errorf("expected POST got %s", req.Method)
	}
	if req.Header.Get("Authorization") != "token" {
		t.Errorf("missing auth header")
	}
}

func TestSendRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	svc := &GroupMeService{Client: server.Client(), Config: &config.GroupMeConfig{}}
	u, _ := url.Parse(server.URL)
	req := &http.Request{Method: "POST", URL: u}
	ok, err := svc.sendRequest(req)
	if err != nil || !ok {
		t.Fatalf("expected success, got %v %v", ok, err)
	}
}

func TestSendRequestFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("fail"))
	}))
	defer server.Close()

	svc := &GroupMeService{Client: server.Client(), Config: &config.GroupMeConfig{}}
	u, _ := url.Parse(server.URL)
	req := &http.Request{Method: "POST", URL: u}
	ok, err := svc.sendRequest(req)
	if err == nil || ok {
		t.Fatalf("expected failure")
	}
}
