package main

import (
	"crowfather/internal/config"
	"crowfather/internal/groupme"
	"crowfather/internal/open_ai"
	"crowfather/internal/router"
	"fmt"
	"github.com/gin-gonic/gin"
)

func main() {
	//db := database.ConnectDb();
	config, err := config.LoadConfig()

	if err != nil {
		fmt.Println("Failed to load config %v", err)
		return
	}

	oai := open_ai.NewOpenAIService(config.OpenAI)
	gms := groupme.NewGroupMeService(config.GroupMe)
	router, err := router.NewRouter(oai, gms, config.Auth)

	if err != nil {
		return
	}

	r := gin.Default()

	router.RegisterRoutes(r)

	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
