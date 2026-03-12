package main

import (
	"context"
	"crowfather/internal/config"
	"crowfather/internal/database"
	"crowfather/internal/groupme"
	"crowfather/internal/open_ai"
	"crowfather/internal/router"
	"fmt"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.LoadConfig()

	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		return
	}

	var repo open_ai.ThreadRepository
	dbSvc, err := database.ConnectDb()
	if err != nil {
		fmt.Printf("Database unavailable, running in memory-only mode: %v\n", err)
	} else {
		pgRepo := database.NewPgThreadRepository(dbSvc.DB())
		if err := pgRepo.Migrate(context.Background()); err != nil {
			fmt.Printf("Schema migration failed, running in memory-only mode: %v\n", err)
		} else {
			repo = pgRepo
		}
	}

	oai := open_ai.NewOpenAIService(cfg.OpenAI, repo)
	gms := groupme.NewGroupMeService(cfg.GroupMe)
	r, err := router.NewRouter(oai, gms, cfg)

	if err != nil {
		return
	}

	engine := gin.Default()
	r.RegisterRoutes(engine)
	engine.Run()
}
