package main

import (
	app "github.com/fatkulllin/metrilo/internal/app/server"
	config "github.com/fatkulllin/metrilo/internal/config/server"
	"github.com/fatkulllin/metrilo/internal/logger"
)

func main() {
	logger.Initialize("INFO")
	config := config.LoadConfig()

	app := app.NewApp(config)
	app.Run()
}
