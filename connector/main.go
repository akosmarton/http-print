package main

import (
	"bufio"
	"crypto/tls"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

type appConfig struct {
	API struct {
		URL string
		Key string
	}
	Printer struct {
		Name        string
		Type        string
		Destination string
	}
	PollingInterval int
}

var config *appConfig

func main() {
	var err error

	log.Println("HTTP Print Connector started")

	config, err = loadAppConfig("config.json")
	if err != nil {
		log.Fatal(err)
	}

	for {
		i, m := pull()
		switch i {
		case 200:
			continue
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

func pull() (int, string) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{},
	}
	client := &http.Client{Transport: tr}
	req, err := http.NewRequest("GET", config.API.URL+"/printers/"+config.Printer.Name+"/jobs/", nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Add("Authorization", "Bearer "+config.API.Key)
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200:
	default:
		return resp.StatusCode, resp.Status
	}

	switch config.Printer.Type {
	case "file":
		f, err := os.OpenFile(config.Printer.Destination, os.O_WRONLY, 0)
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
	case "tcp":
		conn, _ := net.Dial("tcp", config.Printer.Destination)
		io.Copy(conn, resp.Body)
		conn.Close()
	}

	return resp.StatusCode, resp.Status
}
