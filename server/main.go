package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var jobQueue *JobQueue

func main() {
	jobQueue = NewJobQueue(100)

	printers := os.Getenv("PRINTERS")
	if printers == "" {
		fmt.Println("PRINTERS environment variable is not set. Exiting.")
		os.Exit(1)
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		fmt.Println("API_KEY environment variable is not set. Exiting.")
		os.Exit(1)
	}

	for printer := range strings.SplitSeq(printers, " ") {
		jobQueue.Init(printer)
		fmt.Printf("Initialized printer queue: %s\n", printer)
	}

	e := echo.New()
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Skipper: func(c echo.Context) bool {
			return c.Path() == "/health"
		},
	}))
	e.Use(middleware.Recover())

	e.GET("/health", healthHandler)
	api := e.Group("/api", middleware.KeyAuth(func(key string, c echo.Context) (bool, error) {
		return key == apiKey, nil
	}))
	api.GET("/printers/:printer", getJob)
	api.POST("/printers/:printer", postJob)
	api.DELETE("/printers/:printer", deleteJobs)

	go func() {
		if err := e.Start(":" + port); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatal(err)
		}
	}()

	q := make(chan os.Signal, 1)
	signal.Notify(q, os.Interrupt, syscall.SIGTERM)
	<-q
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
}
