package main

import (
	"os"

	app "github.com/fatkulllin/metrilo/internal/app/agent"
	config "github.com/fatkulllin/metrilo/internal/config/agent"
)

func main() {
	config := config.LoadConfig()
	app := app.NewApp(config)
	if err := app.Run(); err != nil {
		os.Exit(1)
	}
}
