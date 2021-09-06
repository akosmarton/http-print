package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/boltdb/bolt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var apikey string
var dbpath string

var db *bolt.DB

func main() {
	var err error

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	apikey = os.Getenv("APIKEY")
	if apikey == "" {
		apikey = "secret"
	}
	dbpath = os.Getenv("DBPATH")
	if dbpath == "" {
		dbpath = "jobs.db"
	}

	e := echo.New()
	e.Logger.SetHeader("${time_rfc3339}\t${level} ${short_file}:${line}")

	db, err = bolt.Open(dbpath, 0600, nil)
	if err != nil {
		e.Logger.Fatal(err)
	}
	defer db.Close()

	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Skipper: func(c echo.Context) bool {
			return c.Path() == "/health"
		},
		Format: "${time_rfc3339}\tHTTP ${remote_ip} ${host} ${method} ${uri} ${status} ${bytes_out} ${latency_human} ${error}\n",
	}))
	e.Use(middleware.Recover())

	if e.Renderer, err = NewTemplate(); err != nil {
		e.Logger.Fatal(err)
	}

	e.GET("/health", healthHandler)
	e.GET("/printers/:printer/jobs/", webHandler)
	api := e.Group("/api", middleware.KeyAuth(func(key string, c echo.Context) (bool, error) {
		return key == apikey, nil
	}))
	api.GET("/printers/:printer/jobs/", getHandler)
	api.POST("/printers/:printer/jobs/", postHandler)

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
