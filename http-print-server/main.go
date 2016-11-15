package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/boltdb/bolt"
	"github.com/facebookgo/grace/gracehttp"
	"github.com/gorilla/mux"
)

type appConfig struct {
	ListenAddress string
	APIKey        string
	JWT           struct {
		Secret string
		Expiry int
	}
	DBPath string
}

var config *appConfig
var printers map[string]interface{}

func itob(v uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}

func btoi(b []byte) uint64 {
	return binary.BigEndian.Uint64(b)
}

var db *bolt.DB

func main() {
	var err error

	config, err = loadAppConfig("config.json")
	if err != nil {
		log.Fatal(err)
	}

	db, err = bolt.Open("jobs.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	signalChan := make(chan os.Signal)
	signal.Notify(signalChan, syscall.SIGUSR2)
	go func() {
		<-signalChan
		db.Close()
	}()

	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("jobs"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})

	r := mux.NewRouter()
	r.HandleFunc("/", webRoot)

	s := r.PathPrefix("/api").Subrouter()
	s.Handle("/login", jsonMiddleware(http.HandlerFunc(apiLogin)))
	s.Handle("/submit", authMiddleware(jsonMiddleware(http.HandlerFunc(apiSubmit))))
	s.Handle("/fetch", authMiddleware(http.HandlerFunc(apiFetch)))

	log.Println("HTTP Print Server started")
	gracehttp.Serve(&http.Server{Addr: config.ListenAddress, Handler: r})
}

func loadAppConfig(filename string) (*appConfig, error) {
	var c appConfig
	fc, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(fc, &c)
	if err != nil {
		return nil, err
	}

	return &c, nil
}
