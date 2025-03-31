package server

import (
	"fmt"
	"net/http"

	"github.com/fatkulllin/metrilo/internal/handlers"
)

type Server struct{}

func NewServer() *Server {
	fmt.Println("Initializing server...")
	return &Server{}
}

func (server *Server) Start() *http.ServeMux {
	fmt.Println("Start server...")
	mux := http.NewServeMux()
	mux.HandleFunc(`/update/{type}/{name}/{value}`, handlers.SaveMetrics)
	return mux
}
