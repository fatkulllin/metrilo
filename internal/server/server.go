package server

import (
	"fmt"
	"net/http"

	"github.com/fatkulllin/metrilo/internal/handlers"
	"github.com/go-chi/chi"
)

type Server struct {
	Address *string
}

func NewServer(address *string) *Server {
	fmt.Println("Initializing server...")
	return &Server{
		Address: address,
	}
}

func (server *Server) Router() *chi.Mux {
	r := chi.NewRouter()
	r.Post(`/update/{type}/{name}/{value}`, handlers.SaveMetrics)
	r.Get(`/value/{type}/{name}`, handlers.GetMetric)
	r.Get("/", handlers.AllGetMetrics)
	return r
}

func (server *Server) Start() {
	fmt.Printf("Start server on %s...", *server.Address)
	http.ListenAndServe(*server.Address, server.Router())
}
