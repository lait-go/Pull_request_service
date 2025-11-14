package http_server

import (
	"net/http"
	"revivers/pkg/logger"
	"time"
)

type Config struct {
	Port    string `default:":8081" envconfig:"HTTP_PORT"`
	Swagger string `envconfig:"SWAG_URL"`
}

type Server struct {
	HTTPServer *http.Server
}

func New(r http.Handler, c Config) *Server {
	r = http.TimeoutHandler(r, time.Second*5, "request timeout")

	h := &Server{
		HTTPServer: &http.Server{
			Addr:    c.Port,
			Handler: r,
		},
	}

	return h
}

func (s *Server) Run() error {
	logger.Info("server listening on ", s.HTTPServer.Addr)
	return s.HTTPServer.ListenAndServe()
}

func (s *Server) Close() {
	if err := s.HTTPServer.Close(); err != nil {
		logger.Error("failed to close server:", err)
	}
}
