package main

import (
	"bytes"
	"encoding/gob"
)

type job struct {
	Timestamp   uint64
	ContentType string
	Len         uint64
	Payload     []byte
}

func (d *job) GobEncode() ([]byte, error) {
	w := new(bytes.Buffer)
	encoder := gob.NewEncoder(w)
	err := encoder.Encode(d.Timestamp)
	if err != nil {
		return nil, err
	}
	err = encoder.Encode(d.ContentType)
	if err != nil {
		return nil, err
	}
	err = encoder.Encode(d.Len)
	if err != nil {
		return nil, err
	}
	err = encoder.Encode(d.Payload)
	if err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

func (d *job) GobDecode(buf []byte) error {
	r := bytes.NewBuffer(buf)
	decoder := gob.NewDecoder(r)
	err := decoder.Decode(&d.Timestamp)
	if err != nil {
		return err
	}
	err = decoder.Decode(&d.ContentType)
	if err != nil {
		return err
	}
	err = decoder.Decode(&d.Len)
	if err != nil {
		return err
	}
	return decoder.Decode(&d.Payload)
}
