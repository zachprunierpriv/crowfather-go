package open_ai

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockThreadRepo struct {
	mu        sync.Mutex
	data      map[string]string
	getErr    error
	saveErr   error
	getCalls  int
	saveCalls int
}

func (m *mockThreadRepo) GetThreadID(_ context.Context, contextID string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.getCalls++
	if m.getErr != nil {
		return "", m.getErr
	}
	return m.data[contextID], nil
}

func (m *mockThreadRepo) SaveThreadID(_ context.Context, contextID, threadID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.saveCalls++
	if m.saveErr != nil {
		return m.saveErr
	}
	m.data[contextID] = threadID
	return nil
}

func newThreadServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"id":"thread_new","object":"thread","created_at":0,"metadata":{},"tool_resources":{}}`)
	}))
}

// Cache hit: repo is never called when the thread is already in the in-memory map.
func TestGetOrCreateThread_CacheHit(t *testing.T) {
	repo := &mockThreadRepo{data: make(map[string]string)}
	svc := &OpenAIService{
		ThreadIds: map[string]string{"group1": "thread_cached"},
		Repo:      repo,
	}

	id, err := svc.GetOrCreateThread("group1")
	require.NoError(t, err)
	assert.Equal(t, "thread_cached", id)
	assert.Equal(t, 0, repo.getCalls)
	assert.Equal(t, 0, repo.saveCalls)
}

// DB hit: thread absent from cache but found in DB; cache is warmed.
func TestGetOrCreateThread_DBHit(t *testing.T) {
	repo := &mockThreadRepo{data: map[string]string{"group1": "thread_from_db"}}
	svc := &OpenAIService{
		ThreadIds: make(map[string]string),
		Repo:      repo,
	}

	id, err := svc.GetOrCreateThread("group1")
	require.NoError(t, err)
	assert.Equal(t, "thread_from_db", id)
	assert.Equal(t, 1, repo.getCalls)
	assert.Equal(t, 0, repo.saveCalls)
	assert.Equal(t, "thread_from_db", svc.ThreadIds["group1"])
}

// Nil repo: no panic when DB is not configured.
func TestGetOrCreateThread_NilRepo(t *testing.T) {
	server := newThreadServer(t)
	defer server.Close()

	svc := newTestService(t, server.URL)
	svc.Repo = nil

	id, err := svc.GetOrCreateThread("group_no_db")
	require.NoError(t, err)
	assert.NotEmpty(t, id)
}

// DB GetThreadID error: service falls back to creating a new thread rather than returning an error.
func TestGetOrCreateThread_DBGetError_FallsBack(t *testing.T) {
	server := newThreadServer(t)
	defer server.Close()

	repo := &mockThreadRepo{data: make(map[string]string), getErr: errors.New("db down")}
	svc := newTestService(t, server.URL)
	svc.Repo = repo

	id, err := svc.GetOrCreateThread("group1")
	require.NoError(t, err)
	assert.NotEmpty(t, id)
}

// DB SaveThreadID error: thread is still returned and cached in-memory.
func TestGetOrCreateThread_DBSaveError_StillReturns(t *testing.T) {
	server := newThreadServer(t)
	defer server.Close()

	repo := &mockThreadRepo{data: make(map[string]string), saveErr: errors.New("db write failed")}
	svc := newTestService(t, server.URL)
	svc.Repo = repo

	id, err := svc.GetOrCreateThread("group1")
	require.NoError(t, err)
	assert.NotEmpty(t, id)
	assert.Equal(t, id, svc.ThreadIds["group1"])
}

// Concurrent calls: no data races (run with go test -race ./internal/open_ai).
func TestGetOrCreateThread_Race(t *testing.T) {
	server := newThreadServer(t)
	defer server.Close()

	repo := &mockThreadRepo{data: make(map[string]string)}
	svc := newTestService(t, server.URL)
	svc.Repo = repo

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = svc.GetOrCreateThread("shared_group")
		}()
	}
	wg.Wait()
}
