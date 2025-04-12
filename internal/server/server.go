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

	r.Route("/update", func(r chi.Router) {
		r.Use(
			common.SetHeaderTextMiddleware,
			common.MethodPostOnlyMiddleware,
		)
		r.With(common.ValidateURLParamsMiddleware,
			common.ValidateTypeMetricMiddleware,
			common.CheckReqHeaderMiddleware).Post("/{type}/{name}/{value}", server.handlers.SaveMetrics)
	})

	r.Route("/value", func(r chi.Router) {
		r.Use(common.SetHeaderTextMiddleware, common.MethodGetOnlyMiddleware)
		r.With(common.ValidateTypeMetricMiddleware).Get("/{type}/{name}", server.handlers.GetMetric)
	})

	r.Group(func(r chi.Router) {
		r.Use(common.SetHeaderHTMLMiddleware, common.MethodGetOnlyMiddleware)
		r.Get("/", server.handlers.AllGetMetrics)
	})

	err := http.ListenAndServe(server.Address, r)
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
