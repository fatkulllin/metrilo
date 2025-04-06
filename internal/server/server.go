package server

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"

	"github.com/fatkulllin/metrilo/internal/handlers"
	"github.com/go-chi/chi"
	"github.com/spf13/pflag"
)

type NetAddress struct {
	Host string
	Port int
}

func (a *NetAddress) Type() string {
	return "netAddress"
}

func (a NetAddress) String() string {
	return a.Host + ":" + strconv.Itoa(a.Port)
}

func (a *NetAddress) Set(s string) error {
	host, portStr, err := net.SplitHostPort(s)
	if err != nil {
		return errors.New("need address in the form host:port")
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return err
	}
	a.Host = host
	a.Port = port
	return nil
}

type Server struct {
	Address NetAddress
}

func NewServer() *Server {
	fmt.Println("Initializing server...")
	return &Server{
		Address: NetAddress{
			Host: "localhost",
			Port: 8080,
		},
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
	pflag.VarP(&server.Address, "address", "a", "localhost:8080")
	pflag.Parse()
	bindAddress := server.Address.String()
	log.Printf("Server started on %s...", bindAddress)
	err := http.ListenAndServe(bindAddress, server.Router())
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
