package http_server

import (
	"net/http"
	"revivers/pkg/logger"
	"time"
)

type Config struct {
	Url    string `default:"0.0.0.0:8080" envconfig:"URL"`
}

type Server struct {
	HTTPServer *http.Server
}

func New(r http.Handler, c Config) *Server {
	r = http.TimeoutHandler(r, time.Second*5, "request timeout")

	return &Server{
			HTTPServer: &http.Server{
				Addr:    c.Url,
				Handler: r,
			},
		}
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
