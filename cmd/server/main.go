package main

import (
	app "github.com/fatkulllin/metrilo/internal/app/server"
	config "github.com/fatkulllin/metrilo/internal/config/server"
)

func main() {
	config := config.LoadConfig()
	app := app.NewApp(config)
	app.Run()
}
