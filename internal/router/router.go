package router

import (
	"crowfather/internal/groupme"
	"crowfather/internal/handlers/message_handler"
	"crowfather/internal/handlers/test_handler"
	"crowfather/internal/open_ai"
	"github.com/gin-gonic/gin"
	"net/http"
)

type Router struct {
	oai            *open_ai.OpenAIService
	gms            *groupme.GroupMeService
	messageHandler func(groupme.Message, *open_ai.OpenAIService, *groupme.GroupMeService) (string, error)
	testHandler    func(string, *open_ai.OpenAIService) (string, error)
}

func NewRouter(oai *open_ai.OpenAIService, gms *groupme.GroupMeService) (*Router, error) {
	messageHandler := message_handler.Handle
	testHandler := test_handler.Handle

	return &Router{
		messageHandler: messageHandler,
		testHandler:    testHandler,
		oai:            oai,
		gms:            gms,
	}, nil
}

func (r *Router) RegisterRoutes(engine *gin.Engine) {
	engine.GET("/ping", r.handlePing)
	engine.POST("/message", r.processGroupMeMessage)
	engine.POST("/test", r.processTestMessage)
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
	response, err := r.messageHandler(groupmeMessage, r.oai, r.gms)

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

	response, err := r.testHandler(message.Text, r.oai)

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
