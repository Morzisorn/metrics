package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/morzisorn/metrics/internal/server/storage"
	"github.com/stretchr/testify/assert"
)

const (
	host = "http://localhost:8080"
)

func TestUpdateCounterOK(t *testing.T) {
	s := storage.GetStorage()
	url := host + "/update/counter/test/1"
	r := httptest.NewRequest("POST", url, nil)
	r.Header.Set("Content-Type", "text/plain")
	w := httptest.NewRecorder()
	Update(w, r)
	Update(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	v, exist := s.GetMetric("test")
	assert.True(t, exist)
	assert.Equal(t, 2.0, v)
}

func TestUpdateGaugeOK(t *testing.T) {
	s := storage.GetStorage()
	url := host + "/update/gauge/test/2.5"
	r := httptest.NewRequest("POST", url, nil)
	r.Header.Set("Content-Type", "text/plain")
	w := httptest.NewRecorder()

	Update(w, r)
	Update(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	v, exist := s.GetMetric("test")
	assert.True(t, exist)
	assert.Equal(t, 2.5, v)
}

func TestUpdateInvalidPath(t *testing.T) {
	url := host + "/update/counter/test"
	r := httptest.NewRequest("POST", url, nil)
	r.Header.Set("Content-Type", "text/plain")
	w := httptest.NewRecorder()

	Update(w, r)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUpdateInvalidMethod(t *testing.T) {
	url := host + "/update/counter/test/1"
	r := httptest.NewRequest("GET", url, nil)
	r.Header.Set("Content-Type", "text/plain")
	w := httptest.NewRecorder()

	Update(w, r)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestUpdateInvalidContentType(t *testing.T) {
	url := host + "/update/counter/test/1"
	r := httptest.NewRequest("POST", url, nil)
	r.Header.Set("Content-Type", "incorrect")
	w := httptest.NewRecorder()

	Update(w, r)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestUpdateInvalidType(t *testing.T) {
	url := host + "/update/incorrect/test/1"
	r := httptest.NewRequest("POST", url, nil)
	r.Header.Set("Content-Type", "text/plain")
	w := httptest.NewRecorder()

	Update(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateInvalidGaugeValue(t *testing.T) {
	url := host + "/update/gauge/test/incorrect"
	r := httptest.NewRequest("POST", url, nil)
	r.Header.Set("Content-Type", "text/plain")
	w := httptest.NewRecorder()

	Update(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateInvalidCounterValue(t *testing.T) {
	url := host + "/update/counter/test/2.5"
	r := httptest.NewRequest("POST", url, nil)
	r.Header.Set("Content-Type", "text/plain")
	w := httptest.NewRecorder()

	Update(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
