package main

import "time"

type Job struct {
	Timestamp   time.Time
	ContentType string
	Payload     []byte
}
