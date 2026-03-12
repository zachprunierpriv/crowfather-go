package open_ai

import (
	"crowfather/internal/config"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Bug 1: GetThreadId must hold RLock when reading ThreadIds.
// Run with: go test -race ./internal/open_ai
func TestGetThreadIdRace(t *testing.T) {
	svc := &OpenAIService{
		ThreadIds: map[string]string{"group1": "thread_abc"},
	}

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			svc.mu.Lock()
			svc.ThreadIds["group2"] = "thread_xyz"
			svc.mu.Unlock()
		}()
		go func() {
			defer wg.Done()
			_ = svc.GetThreadId("group1")
		}()
	}
	wg.Wait()
}

// Bug 2 helpers.

func runStatusServer(t *testing.T, status string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"id":"run_1","object":"thread.run","created_at":0,"thread_id":"thread_1","assistant_id":"asst_1","status":%q,"model":"gpt-4o","tools":[],"metadata":{},"parallel_tool_calls":true,"response_format":"auto","tool_choice":"auto"}`, status)
	}))
}

func newTestService(t *testing.T, baseURL string) *OpenAIService {
	t.Helper()
	opts := []option.RequestOption{
		option.WithAPIKey("test"),
		option.WithBaseURL(baseURL),
	}
	threadC := openai.NewBetaThreadService(opts[0])
	return &OpenAIService{
		ThreadClient: &threadC,
		Config:       &config.OpenAIConfig{Timeout: 500 * time.Millisecond},
		Options:      opts,
		ThreadIds:    make(map[string]string),
	}
}

// Bug 2a: cancelled run must return an error immediately, not loop until timeout.
func TestGetResponseCancelledReturnsError(t *testing.T) {
	server := runStatusServer(t, "cancelled")
	defer server.Close()

	svc := newTestService(t, server.URL)
	run := openai.Run{ID: "run_1", ThreadID: "thread_1"}

	_, err := svc.GetResponse(run, "msg_1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cancelled")
}

// Bug 2b: requires_action run must return an error immediately, not loop until timeout.
func TestGetResponseRequiresActionReturnsError(t *testing.T) {
	server := runStatusServer(t, "requires_action")
	defer server.Close()

	svc := newTestService(t, server.URL)
	run := openai.Run{ID: "run_1", ThreadID: "thread_1"}

	_, err := svc.GetResponse(run, "msg_1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "requires action")
}
