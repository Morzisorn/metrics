package main

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/morzisorn/metrics/config"
	"github.com/stretchr/testify/require"
)

func TestRunAgent(t *testing.T) {
	Service = config.GetService("agent")
	Service.Config.PollInterval = 1
	Service.Config.ReportInterval = 2

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(Service.Config.ReportInterval*float64(time.Second))+3*time.Second)
	defer cancel()

	server := runMockServerRespondsToUpdate()
	defer server.Close()

	url := strings.Split(server.URL, "://")[1]
	Service.Config.Addr = url 

	errCh := make(chan error, 1)

	go func() {
		errCh <- RunAgent() // функция "виснет", если всё нормально
	}()

	select {
	case <-ctx.Done():
		// Всё хорошо, RunAgent не вернула ошибку за отведённое время
	case err := <-errCh:
		require.NoError(t, err, "RunAgent returned an unexpected error")
	}
}

func runMockServerRespondsToUpdate() *httptest.Server{
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/update/" {
			w.WriteHeader(http.StatusOK)
			_, err := io.WriteString(w, `ok`)
			if err != nil {
				return 
			}
			return
		}
		http.NotFound(w, r)
	}))
}
