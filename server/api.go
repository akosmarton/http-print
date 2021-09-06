package main

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"github.com/boltdb/bolt"
)

func jobPost(w http.ResponseWriter, r *http.Request) {
	var buf bytes.Buffer
	printer := mux.Vars(r)["printer"]
	n, _ := buf.ReadFrom(r.Body)

	j := job{
		Timestamp:   uint64(time.Now().Unix()),
		ContentType: r.Header.Get("Content-type"),
		Len:         uint64(r.ContentLength),
		Payload:     buf.Bytes(),
	}

	db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(printer))
		if b == nil {
			var err error
			b, err = tx.CreateBucket([]byte(printer))
			if err != nil {
				http.Error(w, "500 Internal Server Error", 500)
				log.Println(err)
				return err
			}
		}

		je, _ := j.GobEncode()
		k := fmt.Sprintf("%016x", time.Now().UnixNano())
		err := b.Put([]byte(k), je)
		if err != nil {
			return err
		}

		remote := r.RemoteAddr
		xff := r.Header.Get("X-Forwarded-For")
		if xff != "" {
			remote = xff
		}
		log.Println(remote, "Job submitted:", n, "byte(s)")
		return nil
	})
}

func jobGet(w http.ResponseWriter, r *http.Request) {
	printer := mux.Vars(r)["printer"]

	db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(printer))
		if b == nil {
			http.Error(w, "404 Not Found", 404)
			return bolt.ErrBucketNotFound
		}

		k, v := b.Cursor().First()

		if k == nil {
			http.Error(w, "404 Not Found", 404)
			return errors.New("Key not found")
		}

		var j job

		j.GobDecode(v)
		w.Header().Set("Content-Type", j.ContentType)
		w.Write(j.Payload)
		b.Delete(k)

		remote := r.RemoteAddr
		xff := r.Header.Get("X-Forwarded-For")
		if xff != "" {
			remote = xff
		}
		log.Println(remote, "Job fetched:", j.Len, "byte(s)")

		return nil
	})
}
