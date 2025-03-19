package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bilbothegreedy/HNS/internal/config"
	"github.com/rs/zerolog/log"
)

// Server represents the API server
type Server struct {
	server *http.Server
	config *config.Config
}

// NewServer creates a new API server
func NewServer(router http.Handler, config *config.Config) *Server {
	return &Server{
		server: &http.Server{
			Addr:         fmt.Sprintf(":%d", config.Server.Port),
			Handler:      router,
			ReadTimeout:  config.Server.ReadTimeout,
			WriteTimeout: config.Server.WriteTimeout,
		},
		config: config,
	}
}

// Start starts the API server
func (s *Server) Start() error {
	log.Info().Msgf("Starting server on port %d", s.config.Server.Port)
	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the API server
func (s *Server) Shutdown(ctx context.Context) error {
	log.Info().Msg("Shutting down server...")
	return s.server.Shutdown(ctx)
}

// StartWithGracefulShutdown starts the server and sets up graceful shutdown
func (s *Server) StartWithGracefulShutdown(quit <-chan struct{}) error {
	// Start server in a goroutine
	go func() {
		if err := s.Start(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Failed to start server")
		}
	}()

	// Wait for quit signal
	<-quit

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), s.config.Server.ShutdownTimeout)
	defer cancel()

	// Shutdown the server
	if err := s.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	log.Info().Msg("Server exited properly")
	return nil
}
