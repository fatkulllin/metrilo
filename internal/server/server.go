package server

import (
	"fmt"
	"net/http"

	"github.com/fatkulllin/metrilo/internal/handlers"
	"github.com/go-chi/chi"
)

type Server struct{}

func NewServer() *Server {
	fmt.Println("Initializing server...")
	return &Server{}
}

func (server *Server) Router() *chi.Mux {
	r := chi.NewRouter()
	r.Post(`/update/{type}/{name}/{value}`, handlers.SaveMetrics)
	r.Get(`/value/{type}/{name}`, handlers.GetMetric)
	r.Get("/", handlers.AllGetMetrics)
	return r
}

func (server *Server) Start() {
	fmt.Println("Start server...")
	http.ListenAndServe(":8080", server.Router())
}
