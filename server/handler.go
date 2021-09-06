package main

import (
	"bytes"
	"encoding/gob"
	"io"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/boltdb/bolt"
)

func healthHandler(c echo.Context) error {
	c.Response().Header().Set("Cache-Control", "no-store, max-age=0")
	return c.String(http.StatusOK, "OK")
}

func postHandler(c echo.Context) error {
	b, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return err
	}

	j := &Job{
		ContentType: c.Request().Header.Get("Content-Type"),
		Payload:     b,
	}

	return db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(c.Param("printer")))
		if bucket == nil {
			var err error
			bucket, err = tx.CreateBucket([]byte(c.Param("printer")))
			if err != nil {
				return err
			}
		}

		buf := &bytes.Buffer{}
		if err := gob.NewEncoder(buf).Encode(j); err != nil {
			return err
		}
		ts, err := time.Now().UTC().MarshalBinary()
		if err != nil {
			return err
		}
		return bucket.Put(ts, buf.Bytes())
	})
}

func getHandler(c echo.Context) error {
	return db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(c.Param("printer")))
		if bucket == nil {
			return echo.ErrNotFound
		}

		k, v := bucket.Cursor().First()
		if k == nil {
			return echo.ErrNotFound
		}
		defer bucket.Delete(k)

		j := &Job{}
		if err := gob.NewDecoder(bytes.NewBuffer(v)).Decode(j); err != nil {
			return err
		}

		return c.Blob(http.StatusOK, j.ContentType, j.Payload)
	})
}

func webHandler(c echo.Context) error {
	return db.View(func(tx *bolt.Tx) error {
		type dataJob struct {
			Timestamp   string
			ContentType string
			Len         int
		}

		data := struct {
			Printer string
			Jobs    []dataJob
		}{
			Printer: c.Param("printer"),
		}

		b := tx.Bucket([]byte(c.Param("printer")))
		if b != nil {
			b.ForEach(func(k, v []byte) error {
				j := &Job{}
				if err := gob.NewDecoder(bytes.NewBuffer(v)).Decode(j); err != nil {
					return err
				}

				t := &time.Time{}
				if err := t.UnmarshalBinary(k); err != nil {
					return err
				}

				dj := dataJob{
					Timestamp:   t.String(),
					ContentType: j.ContentType,
					Len:         len(j.Payload),
				}

				data.Jobs = append(data.Jobs, dj)

				return nil
			})
		}
		return c.Render(http.StatusOK, "", data)
	})
}
