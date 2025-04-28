package router

import (
	"crowfather/internal/config"
	"crowfather/internal/groupme"
	"crowfather/internal/handlers/meltdown_handler"
	"crowfather/internal/handlers/message_handler"
	"crowfather/internal/handlers/test_handler"
	"crowfather/internal/open_ai"
	"github.com/gin-gonic/gin"
	"net/http"
)

type Router struct {
	oai             *open_ai.OpenAIService
	gms             *groupme.GroupMeService
	messageHandler  func(groupme.Message, *open_ai.OpenAIService, *groupme.GroupMeService, string) (string, error)
	testHandler     func(string, *open_ai.OpenAIService, string) (string, error)
	meltdownHandler func(string, *open_ai.OpenAIService, string) (string, error)
	config          *config.Config
}

func NewRouter(oai *open_ai.OpenAIService, gms *groupme.GroupMeService, config *config.Config) (*Router, error) {
	messageHandler := message_handler.Handle
	testHandler := test_handler.Handle
	meltdownHandler := meltdown_handler.Handle

	return &Router{
		messageHandler:  messageHandler,
		testHandler:     testHandler,
		meltdownHandler: meltdownHandler,
		oai:             oai,
		gms:             gms,
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
}

func (r *Router) handlePing(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
	})
}

func (r *Router) processGroupMeMessage(c *gin.Context) {
	var groupmeMessage groupme.Message

	if err := c.BindJSON(&groupmeMessage); err != nil {
		return
	}
	response, err := r.messageHandler(groupmeMessage, r.oai, r.gms, r.config.Assistants.GroupMeAssistantID)

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
