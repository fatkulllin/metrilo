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

func (server *Server) Start() {
	fmt.Println("Start server...")
	// mux := http.NewServeMux()
	// mux.HandleFunc(`/update/{type}/{name}/{value}`, handlers.SaveMetrics)
	// return mux
	r := chi.NewRouter()
	r.Post(`/update/{type}/{name}/{value}`, handlers.SaveMetrics)
	// r.Get("/item/{id}", func(rw http.ResponseWriter, r *http.Request) {
	// 	// получаем значение URL-параметра id
	// 	id := chi.URLParam(r, "id")
	// 	io.WriteString(rw, fmt.Sprintf("item = %s", id))
	// })
	r.Get(`/value/{type}/{name}`, handlers.GetMetric)
	r.Get("/", handlers.AllGetMetrics)
	// r передаётся как http.Handler
	http.ListenAndServe(":8080", r)
}
