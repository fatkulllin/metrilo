package main

import (
	"github.com/fatkulllin/metrilo/internal/server"
	"github.com/spf13/pflag"
)

func main() {
	var address *string = pflag.StringP("address", "a", "localhost:8080", "port start server")
	pflag.Parse()
	server := server.NewServer(address)
	server.Start()

}
