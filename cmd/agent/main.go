package main

import (
	app "github.com/fatkulllin/metrilo/internal/app/agent"
	config "github.com/fatkulllin/metrilo/internal/config/agent"
	"github.com/fatkulllin/metrilo/internal/logger"
)

func main() {
	logger.Initialize("INFO")
	config := config.LoadConfig()
	app := app.NewApp(config)
	app.Run()
}
