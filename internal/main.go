package main

import (
	"crowfather/internal/groupme"
	"crowfather/internal/handlers"
	"crowfather/internal/open_ai"
	"github.com/gin-gonic/gin"
	"net/http"
)

func main() {
	//db := database.ConnectDb();
	oai := open_ai.NewOpenAIService()
	gms := groupme.NewGroupMeService()

	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	r.POST("/message", func(c *gin.Context) {
		var groupmeMessage groupme.Message
		if err := c.BindJSON(&groupmeMessage); err != nil {
			return
		}
		response, err := message.MessageHandler(groupmeMessage, oai, gms)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"response": response,
		})
	})
	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
