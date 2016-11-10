package main

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/akosmarton/http-print/protocol"
)

type appConfig struct {
	API struct {
		URL string
		Key string
	}
	Printer struct {
		Type        string
		Destination string
	}
	PollingInterval int
}

var config *appConfig
var token string

func main() {
	var err error

	log.Println("HTTP Print Connector started")

	config, err = loadAppConfig("config.json")
	if err != nil {
		log.Fatal(err)
	}

	if err := login(); err != nil {
		log.Fatal(err)
	}

	for {
		i, m := pull()
		switch i {
		case 200:
		case 403:
			if err := login(); err != nil {
				log.Fatal(err)
			}
		case 404:
			time.Sleep(time.Second * time.Duration(config.PollingInterval))
		default:
			log.Fatal(m)
		}
	}
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

	c.API.URL = strings.TrimSuffix(c.API.URL, "/")

	return &c, nil
}

func login() error {
	areq := protocol.AuthReq{APIKey: config.API.Key}
	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(areq)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{},
	}
	client := &http.Client{Transport: tr}

	resp, err := client.Post(config.API.URL+"/login", "application/json; charset=utf-8", b)

	if err != nil {
		log.Fatal(err)
	}

	aresp := protocol.AuthResp{}
	json.NewDecoder(resp.Body).Decode(&aresp)

	if aresp.AccessToken == "" {
		return errors.New("Authentication failed")
	}

	token = aresp.AccessToken

	return nil
}

func pull() (int, string) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{},
	}
	client := &http.Client{Transport: tr}
	req, err := http.NewRequest("GET", config.API.URL+"/fetch", nil)
	req.Header.Add("Authorization", "Bearer "+token)
	resp, err := client.Do(req)

	if err != nil {
		log.Fatal(err)
	}

	switch resp.StatusCode {
	case 200:
	default:
		return resp.StatusCode, resp.Status
	}

	switch config.Printer.Type {
	case "file":
		f, err := os.OpenFile(config.Printer.Destination, os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			log.Fatal(err)
		}
		w := bufio.NewWriter(f)
		written, err := io.Copy(w, resp.Body)
		if err != nil {
			f.Close()
			log.Fatal(err)
		}
		w.Flush()
		f.Close()
		log.Println("Written:", written, "byte(s)")
	case "pipe":
		cmd := exec.Command(config.Printer.Destination)
		pipe, _ := cmd.StdinPipe()
		cmd.Start()
		io.Copy(pipe, resp.Body)
		pipe.Close()
		cmd.Wait()
	}

	return resp.StatusCode, resp.Status
}
