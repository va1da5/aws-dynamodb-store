package server

import (
	"aws-dynamodb-store/internal/config"
	"aws-dynamodb-store/internal/repository"
	"aws-dynamodb-store/internal/service"
	"fmt"
	"net/http"
	"time"

	_ "github.com/joho/godotenv/autoload"
)

type Server struct {
	port       int
	repository repository.Repository
	service    *service.Service
}

func NewServer(cfg config.AppConfig, repository repository.Repository, service *service.Service) *http.Server {
	NewServer := &Server{
		port:       cfg.ServerPort,
		repository: repository,
		service:    service,
	}

	// Declare Server config
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", NewServer.port),
		Handler:      NewServer.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server
}
