package cmd

import (
	"testing"
)

func TestGetAvailablePort(t *testing.T) {
	port, err := getAvailablePort()
	if err != nil {
		t.Error(err)
	}
	if port < 1 {
		t.Errorf("generated port should be > 1, got %d", port)
	}
}
