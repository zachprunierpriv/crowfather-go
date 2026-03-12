package reconciler

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newGuardTestReconciler() *Reconciler {
	return &Reconciler{cooldown: 0}
}

func TestTrigger_UnauthorizedUser(t *testing.T) {
	r := newGuardTestReconciler()
	r.approvedUsers = map[string]bool{"allowed": true}

	triggered, reason := r.Trigger("stranger", nil)
	assert.False(t, triggered)
	assert.Contains(t, reason, "not authorized")
}

func TestTrigger_ApprovedUserAllowed(t *testing.T) {
	r := newGuardTestReconciler()
	r.approvedUsers = map[string]bool{"allowed": true}

	done := make(chan string, 1)
	triggered, reason := r.Trigger("allowed", func(s string) { done <- s })
	assert.True(t, triggered)
	assert.Empty(t, reason)
	<-done // wait for goroutine to finish
}

func TestTrigger_EmptySenderBypassesAccessControl(t *testing.T) {
	r := newGuardTestReconciler()
	r.approvedUsers = map[string]bool{"allowed": true}

	done := make(chan string, 1)
	triggered, reason := r.Trigger("", func(s string) { done <- s })
	assert.True(t, triggered)
	assert.Empty(t, reason)
	<-done
}

func TestTrigger_NoApprovedListAllowsAnyone(t *testing.T) {
	r := newGuardTestReconciler()
	// approvedUsers is nil/empty — anyone can trigger

	done := make(chan string, 1)
	triggered, _ := r.Trigger("random-user", func(s string) { done <- s })
	assert.True(t, triggered)
	<-done
}

func TestTrigger_CooldownBlocks(t *testing.T) {
	r := &Reconciler{
		cooldown:  time.Hour,
		lastRunAt: time.Now(),
	}

	triggered, reason := r.Trigger("", nil)
	assert.False(t, triggered)
	assert.Contains(t, reason, "just refreshed")
}

func TestTrigger_InFlightBlocks(t *testing.T) {
	r := newGuardTestReconciler()
	r.running = true

	triggered, reason := r.Trigger("", nil)
	assert.False(t, triggered)
	assert.Contains(t, reason, "already in progress")
}

func TestTrigger_NotifyCalledAfterRun(t *testing.T) {
	r := newGuardTestReconciler()
	done := make(chan string, 1)

	triggered, _ := r.Trigger("", func(s string) { done <- s })
	require.True(t, triggered)

	select {
	case summary := <-done:
		assert.NotEmpty(t, summary)
	case <-time.After(5 * time.Second):
		t.Error("notify was not called within 5s")
	}
}

func TestTrigger_SetsRunningFalseAfterCompletion(t *testing.T) {
	r := newGuardTestReconciler()
	done := make(chan string, 1)

	triggered, _ := r.Trigger("", func(s string) { done <- s })
	require.True(t, triggered)

	<-done

	r.mu.Lock()
	defer r.mu.Unlock()
	assert.False(t, r.running)
	assert.False(t, r.lastRunAt.IsZero())
}
