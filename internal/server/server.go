package server

import (
	"fmt"
	"log"
	"net/http"

	config "github.com/fatkulllin/metrilo/internal/config/server"
	"github.com/fatkulllin/metrilo/internal/handlers"
	"github.com/fatkulllin/metrilo/internal/middleware/common"
	"github.com/go-chi/chi"
)

type Server struct {
	Address  string
	handlers *handlers.Handlers
}

func NewServer(handlers *handlers.Handlers, cfg *config.Config) *Server {
	fmt.Println("Initializing server...")
	return &Server{
		Address:  cfg.Address,
		handlers: handlers,
	}
}

func (server *Server) Start() {

	log.Printf("Server started on %s...", server.Address)

	r := chi.NewRouter()

	r.Route("/update/{type}/{name}/{value}", func(r chi.Router) {
		r.Use(common.SetHeaderTextMiddleware, common.CheckReqHeaderMiddleware, common.MethodPostOnlyMiddleware, common.ValidateURLParamsMiddleware, common.ValidateTypeMetricMiddleware)
		r.Post("/", server.handlers.SaveMetrics)
	})

	r.Route("/value/{type}/{name}", func(r chi.Router) {
		r.Use(common.SetHeaderTextMiddleware, common.ValidateTypeMetricMiddleware, common.MethodGetOnlyMiddleware)
		r.Get("/", server.handlers.GetMetric)
	})

	r.Route("/", func(r chi.Router) {
		r.Use(common.SetHeaderHTMLMiddleware, common.MethodGetOnlyMiddleware)
		r.Get("/", server.handlers.AllGetMetrics)
	})

	err := http.ListenAndServe(server.Address, r)
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
