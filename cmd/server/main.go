package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"

	app "github.com/fatkulllin/metrilo/internal/app/server"
	config "github.com/fatkulllin/metrilo/internal/config/server"
	"github.com/fatkulllin/metrilo/internal/logger"
	"go.uber.org/zap"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func main() {
	if buildVersion == "" {
		buildVersion = "N/A"
	}
	if buildDate == "" {
		buildDate = "N/A"
	}
	if buildCommit == "" {
		buildCommit = "N/A"
	}
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)

	logger.Initialize("INFO")
	config := config.LoadConfig()

	app := app.NewApp(config)
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()
	if err := app.Run(ctx); err != nil {
		logger.Log.Fatal("app shutdown with error", zap.Error(err))
	}
}
