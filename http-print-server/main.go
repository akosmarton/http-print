package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/boltdb/bolt"
	"github.com/gorilla/mux"
)

var apikey string
var dbpath string

var db *bolt.DB

func main() {
	var err error

	apikey = os.Getenv("APIKEY")
	dbpath = os.Getenv("DBPATH")
	if dbpath == "" {
		dbpath = "jobs.db"
	}

	db, err = bolt.Open(dbpath, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	r := mux.NewRouter()
	r.Methods("OPTIONS").HandlerFunc(optionsHandler)
	r.Path("/printers/{printer}/jobs/").Methods("GET").HandlerFunc(webRoot)
	r.Path("/api/printers/{printer}/jobs/").Methods("POST").HandlerFunc(jobPost)
	r.Path("/api/printers/{printer}/jobs/").Methods("GET").HandlerFunc(jobGet)
	r.Use(corsMiddleware, authMiddleware)
	log.Println("HTTP Print Server started")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: r,
	}

	signalChan := make(chan os.Signal)
	signal.Notify(signalChan, syscall.SIGUSR2)
	go func() {
		<-signalChan
		db.Close()
	}()

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err.Error())
	}
}

func optionsHandler(w http.ResponseWriter, r *http.Request) {
}
