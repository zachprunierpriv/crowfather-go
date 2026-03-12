package router

import (
	"crowfather/internal/config"
	"crowfather/internal/groupme"
	"crowfather/internal/handlers/meltdown_handler"
	"crowfather/internal/handlers/message_handler"
	"crowfather/internal/handlers/test_handler"
	"crowfather/internal/open_ai"
	"crowfather/internal/reconciler"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type Router struct {
	oai             *open_ai.OpenAIService
	gms             *groupme.GroupMeService
	rec             *reconciler.Reconciler // nil if reconciler is not configured
	messageHandler  func(groupme.Message, *open_ai.OpenAIService, *groupme.GroupMeService, string) (string, error)
	testHandler     func(string, *open_ai.OpenAIService, string) (string, error)
	meltdownHandler func(string, *open_ai.OpenAIService, string) (string, error)
	config          *config.Config
}

func NewRouter(oai *open_ai.OpenAIService, gms *groupme.GroupMeService, rec *reconciler.Reconciler, config *config.Config) (*Router, error) {
	return &Router{
		messageHandler:  message_handler.Handle,
		testHandler:     test_handler.Handle,
		meltdownHandler: meltdown_handler.Handle,
		oai:             oai,
		gms:             gms,
		rec:             rec,
		config:          config,
	}, nil
}

func (r *Router) RegisterRoutes(engine *gin.Engine) {
	engine.GET("/ping", r.handlePing)
	engine.POST("/message", r.processGroupMeMessage)
	engine.POST("/meltdown", r.processMeltdownMessage)

	base := engine.Group("/")
	base.Use(AuthMiddleware(r.config.Auth.APIKey))
	base.POST("/test", r.processTestMessage)
	base.POST("/refresh", r.handleRefresh)
}

func (r *Router) handlePing(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
	})
}

func (r *Router) processGroupMeMessage(c *gin.Context) {
	var msg groupme.Message

	if err := c.BindJSON(&msg); err != nil {
		return
	}

	// Check for the GroupMe refresh trigger before routing to OpenAI.
	if r.rec != nil && isRefreshTrigger(msg.Text) {
		reply := r.handleGroupMeRefresh(msg)
		if err := r.gms.SendRawMessage(reply); err != nil {
			fmt.Printf("router: failed to send refresh reply: %v\n", err)
		}
		c.JSON(http.StatusOK, gin.H{})
		return
	}

	response, err := r.messageHandler(msg, r.oai, r.gms, r.config.Assistants.GroupMeAssistantID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"response": response,
	})
}

func (r *Router) processTestMessage(c *gin.Context) {
	var message struct {
		Text string `json:"text"`
	}

	if err := c.BindJSON(&message); err != nil {
		return
	}

	response, err := r.testHandler(message.Text, r.oai, r.config.Assistants.TestAssistantID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"response": response,
	})
}

func (r *Router) processMeltdownMessage(c *gin.Context) {
	var message struct {
		Text string `json:"text"`
	}

	if err := c.BindJSON(&message); err != nil {
		return
	}

	response, err := r.meltdownHandler(message.Text, r.oai, r.config.Assistants.MeltdownAssistantID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"response": response,
	})
}

// handleRefresh is the HTTP-triggered reconciliation endpoint (POST /refresh).
// Protected by API key middleware. Returns 202 immediately; run is asynchronous.
func (r *Router) handleRefresh(c *gin.Context) {
	if r.rec == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "reconciler not configured"})
		return
	}

	triggered, reason := r.rec.Trigger("", nil)
	if !triggered {
		c.JSON(http.StatusConflict, gin.H{"error": reason})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"status": "refresh started"})
}

// handleGroupMeRefresh processes the GroupMe refresh trigger keyword.
// It sends an immediate acknowledgement and the notify callback posts the result.
func (r *Router) handleGroupMeRefresh(msg groupme.Message) string {
	notify := func(summary string) {
		if err := r.gms.SendRawMessage(summary); err != nil {
			fmt.Printf("router: failed to send refresh summary: %v\n", err)
		}
	}

	triggered, reason := r.rec.Trigger(msg.UserId, notify)
	if triggered {
		return fmt.Sprintf("@%s On it! I'll post an update when the roster refresh is done.", msg.Name)
	}
	return fmt.Sprintf("@%s %s", msg.Name, reason)
}

// isRefreshTrigger returns true if the message text contains the refresh keyword.
func isRefreshTrigger(text string) bool {
	return strings.Contains(strings.ToLower(text), "hey crowfather refresh")
}

func AuthMiddleware(apiKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.GetHeader("Authorization")

		if key == "" {
			key = c.Query("api_key")
		}

		if key == "" || key != apiKey {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "unauthorized",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
