package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/akosmarton/http-print/protocol"
	"github.com/akosmarton/simplejwt"
	"github.com/akosmarton/umid"

	"github.com/boltdb/bolt"
)

var apiState struct {
	sync.Mutex
	LastFetchTime time.Time
}

func apiRoot(w http.ResponseWriter, r *http.Request) {
	j := json.NewEncoder(w)
	j.Encode(&printers)
}

func apiLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "405 Method Not Allowed ("+r.Method+")", 405)
		log.Println(r.RemoteAddr, "405 Method Not Allowed ("+r.Method+")")
		return
	}

	jdec := json.NewDecoder(r.Body)

	req := protocol.AuthReq{}

	jdec.Decode(&req)

	if req.APIKey == config.APIKey {
		jenc := json.NewEncoder(w)

		resp := new(protocol.AuthResp)

		f := make(map[string]interface{})
		f["iat]"] = time.Now().UTC().Unix()
		f["exp"] = time.Now().Add(time.Duration(config.JWT.Expiry) * time.Second).UTC().Unix()
		f["nbf"] = time.Now().Add(-600 * time.Second).UTC().Unix()

		t, _ := simplejwt.NewToken(f, []byte(config.JWT.Secret))

		resp.AccessToken = t

		jenc.Encode(&resp)
		r.Body.Close()
	} else {
		http.Error(w, "401 Unauthorized (Invalid credentials)", 401)
		return
	}
}

func apiSubmit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "405 Method Not Allowed ("+r.Method+")", 405)
		log.Println(r.RemoteAddr, "405 Method Not Allowed ("+r.Method+")")
		return
	}

	var buf bytes.Buffer
	n, _ := buf.ReadFrom(r.Body)

	log.Println(r.RemoteAddr, "Job submitted:", n, "byte(s)")

	j := job{
		Timestamp:   uint64(time.Now().Unix()),
		ContentType: r.Header.Get("Content-type"),
		Len:         uint64(r.ContentLength),
		Payload:     buf.Bytes(),
	}

	db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("jobs"))

		je, _ := j.GobEncode()
		err := b.Put(umid.NewAsBytes(), je)

		return err
	})
}

func apiFetch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "405 Method Not Allowed ("+r.Method+")", 405)
		log.Println(r.RemoteAddr, "405 Method Not Allowed ("+r.Method+")")
		return
	}

	apiState.Lock()
	apiState.LastFetchTime = time.Now().UTC()
	apiState.Unlock()

	db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("jobs"))

		k, v := b.Cursor().First()

		if k == nil {
			http.Error(w, "404 Not Found", 404)
			return nil
		}

		var j job

		j.GobDecode(v)
		w.Header().Set("Content-Type", j.ContentType)
		w.Write(j.Payload)
		b.Delete(k)

		log.Println(r.RemoteAddr, "Job fetched:", j.Len, "byte(s)")

		return nil
	})
}
