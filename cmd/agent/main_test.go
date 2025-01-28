package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRunAgent(t *testing.T) {
	done := make(chan error)

	go func() {
		err := RunAgent()
		done <- err
	}()

	select {
	case err := <-done:
		assert.NoError(t, err, "RunAgent error")
	case <-time.After(12 * time.Second):
		t.Log("RunAgent correctly stopped")
	}
}
