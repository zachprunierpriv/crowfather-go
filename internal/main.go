package main

import (
	"context"
	"crowfather/internal/config"
	"crowfather/internal/database"
	"crowfather/internal/espn"
	"crowfather/internal/groupme"
	"crowfather/internal/open_ai"
	"crowfather/internal/reconciler"
	"crowfather/internal/router"
	"crowfather/internal/sleeper"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		return
	}

	// Database — optional. Service runs in memory-only mode if unavailable.
	var threadRepo open_ai.ThreadRepository
	var metaRepo reconciler.MetadataRepository
	dbSvc, err := database.ConnectDb()
	if err != nil {
		fmt.Printf("Database unavailable, running in memory-only mode: %v\n", err)
	} else {
		pgThread := database.NewPgThreadRepository(dbSvc.DB())
		if err := pgThread.Migrate(context.Background()); err != nil {
			fmt.Printf("Thread schema migration failed: %v\n", err)
		} else {
			threadRepo = pgThread
		}

		pgMeta := database.NewPgMetadataRepository(dbSvc.DB())
		if err := pgMeta.Migrate(context.Background()); err != nil {
			fmt.Printf("Metadata schema migration failed: %v\n", err)
		} else {
			metaRepo = pgMeta
		}
	}

	oai := open_ai.NewOpenAIService(cfg.OpenAI, threadRepo)
	gms := groupme.NewGroupMeService(cfg.GroupMe)

	// Reconciler — optional. Only constructed when SLEEPER_LEAGUE_IDS is set.
	var rec *reconciler.Reconciler
	if cfg.Reconciler != nil {
		rec = reconciler.NewReconciler(
			espn.NewESPNService(),
			sleeper.NewSleeperService(),
			oai,
			metaRepo,
			cfg.Reconciler.LeagueIDs,
			cfg.Assistants.GroupMeAssistantID,
			cfg.Reconciler.TransactionRounds,
			cfg.Reconciler.CooldownMinutes,
			cfg.Reconciler.ApprovedUsers,
		)

		// Startup trigger.
		if cfg.Reconciler.OnStartup {
			fmt.Println("Starting initial roster reconciliation...")
			rec.Trigger("", nil)
		}

		// Periodic cron goroutine.
		go func() {
			ticker := time.NewTicker(cfg.Reconciler.Interval)
			defer ticker.Stop()
			for range ticker.C {
				fmt.Println("Cron: triggering scheduled roster reconciliation")
				rec.Trigger("", nil)
			}
		}()
	}

	r, err := router.NewRouter(oai, gms, rec, cfg)
	if err != nil {
		return
	}

	engine := gin.Default()
	r.RegisterRoutes(engine)
	engine.Run()
}
