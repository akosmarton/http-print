package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/labstack/echo/v4"
)

func healthHandler(c echo.Context) error {
	c.Response().Header().Set("Cache-Control", "no-store, max-age=0")
	return c.String(http.StatusOK, "OK")
}

func postJob(c echo.Context) error {
	b, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return err
	}

	jobs := jobQueue.Get(c.Param("printer"))
	if jobs == nil {
		return c.String(http.StatusNotFound, "Printer not found")
	}

	select {
	case jobs <- Job{
		ContentType: c.Request().Header.Get("Content-Type"),
		Payload:     b,
	}:
	default:
		return fmt.Errorf("Job queue full")
	}

	return c.NoContent(http.StatusCreated)
}

func getJob(c echo.Context) error {
	c.Response().Header().Set("Cache-Control", "no-cache")
	c.Response().Header().Set("Connection", "keep-alive")
	c.Response().Flush()

	jobs := jobQueue.Get(c.Param("printer"))
	if jobs == nil {
		return c.String(http.StatusNotFound, "Printer not found")
	}

	select {
	case <-c.Request().Context().Done():
		return nil
	case j := <-jobs:
		return c.Blob(http.StatusOK, j.ContentType, j.Payload)
	}
}

func deleteJobs(c echo.Context) error {
	if jobQueue.Get(c.Param("printer")) == nil {
		return c.String(http.StatusNotFound, "Printer not found")
	}
	jobQueue.Clear(c.Param("printer"))
	return c.NoContent(http.StatusNoContent)
}
