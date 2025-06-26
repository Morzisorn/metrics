package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/morzisorn/metrics/config"
	"github.com/stretchr/testify/require"
)

func TestRunAgent(t *testing.T) {
	Service = config.GetService("agent")
	Service.Config.PollInterval = 1
	Service.Config.ReportInterval = 2
	Service.Config.Protocol = "http"

	server := runMockServerRespondsToUpdate()
	defer server.Close()

	url := strings.Split(server.URL, "://")[1]
	Service.Config.Addr = url

	errCh := make(chan error, 1)

	go func() {
		errCh <- RunAgent()
	}()

	// Ждем несколько циклов отправки
	time.Sleep(5 * time.Second)

	// Отправляем сигнал завершения
	proc, _ := os.FindProcess(os.Getpid())
	err := proc.Signal(syscall.SIGTERM)
	require.NoError(t, err)

	select {
	case err := <-errCh:
		require.NoError(t, err)
	case <-time.After(3 * time.Second):
		t.Fatal("RunAgent did not shut down gracefully")
	}
}

func runMockServerRespondsToUpdate() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && (r.URL.Path == "/update/" || r.URL.Path == "/updates/") {
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
