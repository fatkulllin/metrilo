package server

import (
	"log"
	"net/http"

	config "github.com/fatkulllin/metrilo/internal/config/server"
	"github.com/fatkulllin/metrilo/internal/handlers"
	"github.com/fatkulllin/metrilo/internal/logger"
	"github.com/fatkulllin/metrilo/internal/middleware/common"
	"github.com/fatkulllin/metrilo/internal/middleware/logging"
	"github.com/go-chi/chi"
	"go.uber.org/zap"
)

type Server struct {
	Address  string
	handlers *handlers.Handlers
}

func NewServer(handlers *handlers.Handlers, cfg *config.Config) *Server {
	logger.Log.Info("Initializing server...")
	return &Server{
		Address:  cfg.Address,
		handlers: handlers,
	}
}

func (server *Server) Start() {

	logger.Log.Info("Server started on...", zap.Any("server", server.Address))

	r := chi.NewRouter()
	r.Use(logging.RequestLogger) // logging.ResponseLogger

	r.Route("/update", func(r chi.Router) {
		r.Use(
			common.MethodPostOnlyMiddleware,
		)
		r.Post("/", server.handlers.SaveJSONMetrics)
		r.With(common.SetHeaderTextMiddleware,
			common.ValidateURLParamsMiddleware,
			common.ValidateTypeMetricMiddleware,
			common.CheckReqHeaderMiddleware).Post("/{type}/{name}/{value}", server.handlers.SaveMetrics)
	})

	r.Route("/value", func(r chi.Router) {
		r.Post("/", server.handlers.GetMetricJSON)
		r.With(common.SetHeaderTextMiddleware, common.MethodGetOnlyMiddleware, common.ValidateTypeMetricMiddleware).Get("/{type}/{name}", server.handlers.GetMetric)
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
