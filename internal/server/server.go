package server

import (
	"crypto/rsa"
	"log"
	"net/http"

	_ "net/http/pprof"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"go.uber.org/zap"

	config "github.com/fatkulllin/metrilo/internal/config/server"
	"github.com/fatkulllin/metrilo/internal/encoder"
	"github.com/fatkulllin/metrilo/internal/handlers"
	"github.com/fatkulllin/metrilo/internal/logger"
	"github.com/fatkulllin/metrilo/internal/middleware/common"
	"github.com/fatkulllin/metrilo/internal/middleware/compressor"
	"github.com/fatkulllin/metrilo/internal/middleware/logging"
)

type Server struct {
	Address    string
	handlers   *handlers.Handlers
	config     *config.Config
	privateKey *rsa.PrivateKey
}

func NewServer(handlers *handlers.Handlers, cfg *config.Config, privateKey *rsa.PrivateKey) *Server {
	logger.Log.Info("Initializing server...")
	server := &Server{
		handlers:   handlers,
		config:     cfg,
		privateKey: privateKey,
	}
	logger.Log.Info("Server URL:", zap.String("server", cfg.Address))
	return server
}

func (server *Server) Start() {

	logger.Log.Info("Server started on...", zap.Any("server", server.config.Address))

	r := chi.NewRouter()
	label := []byte("agent")
	r.Use(logging.RequestLogger) // logging.ResponseLogger
	r.Use(common.NewDecodeMsgMiddleware([]byte(server.config.Key), server.config.WasKeySet))
	r.Use(encoder.DecodeMiddleware(server.privateKey, label))
	r.Use(compressor.GzipMiddleware)
	r.Mount("/debug", middleware.Profiler())
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

	r.Get("/ping", server.handlers.PingDatabase)
	r.Post("/updates/", server.handlers.UpdateMetrics)

	err := http.ListenAndServe(server.config.Address, r)
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
