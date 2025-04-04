package main

import (
	"github.com/fatkulllin/metrilo/internal/server"
)

func main() {
	server := server.NewServer()
	server.Start()

}
