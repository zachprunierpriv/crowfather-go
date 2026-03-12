package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestIsRefreshTrigger(t *testing.T) {
	cases := []struct {
		text    string
		want    bool
	}{
		{"hey crowfather refresh", true},
		{"HEY CROWFATHER REFRESH", true},
		{"Hey Crowfather Refresh the rosters", true},
		{"please hey crowfather refresh now", true},
		{"hey crowfather, what's up?", false},
		{"refresh the rosters", false},
		{"", false},
	}
	for _, tc := range cases {
		assert.Equal(t, tc.want, isRefreshTrigger(tc.text), "input: %q", tc.text)
	}
}

func TestHandleRefresh_NilReconciler_Returns503(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodPost, "/refresh", nil)

	r := &Router{rec: nil}
	r.handleRefresh(c)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}
