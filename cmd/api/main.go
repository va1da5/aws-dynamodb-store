package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"aws-dynamodb-store/internal/config"
	"aws-dynamodb-store/internal/server"
	"aws-dynamodb-store/internal/service"

	dynamodbrepo "aws-dynamodb-store/internal/repository/dynamodb"
)

func gracefulShutdown(apiServer *http.Server, done chan bool) {
	// Create context that listens for the interrupt signal from the OS.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Listen for the interrupt signal.
	<-ctx.Done()

	log.Println("shutting down gracefully, press Ctrl+C again to force")

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := apiServer.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown with error: %v", err)
	}

	log.Println("Server exiting")

	// Notify the main goroutine that the shutdown is complete
	done <- true
}

func main() {

	appCfg, err := config.LoadConfig()

	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Printf("Log level set to: %s", appCfg.LogLevel)
	log.Printf("Server starting on port: %d", appCfg.ServerPort)
	log.Printf("Using DynamoDB table: %s in region: %s", appCfg.DynamoDB.TableName, appCfg.DynamoDB.AWSRegion)
	if appCfg.DynamoDB.UseDynamoDBLocal {
		log.Printf("Connecting to DynamoDB Local at: %s", appCfg.DynamoDB.DynamoDBLocalURL)
	}

	repository := dynamodbrepo.NewDynamoDBRepository(appCfg.DynamoDB)

	services := &service.Service{
		RBACService: service.NewRBACService(repository),
	}

	server := server.NewServer(*appCfg, repository, services)

	// Create a done channel to signal when the shutdown is complete
	done := make(chan bool, 1)

	// Run graceful shutdown in a separate goroutine
	go gracefulShutdown(server, done)

	err = server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		panic(fmt.Sprintf("http server error: %s", err))
	}

	// Wait for the graceful shutdown to complete
	<-done
	log.Println("Graceful shutdown complete.")
}
