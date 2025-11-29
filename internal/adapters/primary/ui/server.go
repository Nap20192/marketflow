package ui

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"
)

type ServerConfig struct {
	Port string
}

func NewServerConfig(port *string) (*ServerConfig, error) {
	p, err := strconv.Atoi(*port)
	if err != nil {
		return nil, err
	} else if p <= 0 || p >= 65000 {
		return nil, err
	}

	return &ServerConfig{
		Port: *port,
	}, nil
}

type Server struct {
	serverConfig *ServerConfig
	server       *http.Server
}

func NewServer(serverConfig *ServerConfig, handler http.Handler) *Server {
	return &Server{
		serverConfig: serverConfig,
		server: &http.Server{
			Addr: ":" + serverConfig.Port,

			Handler: handler,
		},
	}
}

func (s *Server) Start() error {
	err := s.server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		slog.Error("Server error", "error", err)
		return err
	}
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
