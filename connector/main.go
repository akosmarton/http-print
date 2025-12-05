package main

import (
	"bufio"
	"io"
	"log"
	"net"
	"net/url"
	"os"
	"time"

	"resty.dev/v3"
)

func main() {
	log.Println("HTTP Print Connector started")

	apiUrl := os.Getenv("API_URL")
	if apiUrl == "" {
		log.Fatal("API_URL environment variable is not set")
	}

	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		log.Fatal("API_KEY environment variable is not set")
	}

	printer_url := os.Getenv("PRINTER_URL")
	if printer_url == "" {
		log.Fatal("PRINTER_URL environment variable is not set")
	}

	purl, err := url.ParseRequestURI(printer_url)
	if err != nil {
		log.Fatal("Invalid API_URL:", err)
	}

	if purl.Scheme != "file" && purl.Scheme != "tcp" {
		log.Fatal("Unsupported PRINTER_URL scheme:", purl.Scheme)
	}

	client := resty.New()
	defer client.Close()
	client.SetAuthToken(apiKey)

	for {
		log.Println("Waiting for print job...")
		res, err := client.R().Get(apiUrl)
		if err != nil {
			log.Println("Error:", err)
			time.Sleep(time.Second * 5)
			continue
		}
		if res.IsSuccess() {
			if err := print(purl, res.Body); err != nil {
				log.Println("Print error:", err)
			} else {
				log.Println("Print successful")
			}
		} else {
			log.Println("Status:", res.Status())
			time.Sleep(time.Second * 5)
			continue
		}
	}
}

func print(purl *url.URL, payload io.ReadCloser) error {
	switch purl.Scheme {
	case "file":
		f, err := os.OpenFile(purl.Path, os.O_WRONLY, 0)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		w := bufio.NewWriter(f)
		defer w.Flush()
		if _, err := io.Copy(w, payload); err != nil {
			return err
		}
	case "tcp":
		conn, err := net.Dial("tcp", purl.Host)
		if err != nil {
			return err
		}
		defer conn.Close()
		if _, err := io.Copy(conn, payload); err != nil {
			return err
		}
	}
	return nil
}
