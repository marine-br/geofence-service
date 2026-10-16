package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestHTTPMuxServesMetricsAndHealth(t *testing.T) {
	serviceMetrics.ObserveMessage("geofence-test", false, time.Now(), nil)
	server := httptest.NewServer(newHTTPMux(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) }))
	defer server.Close()

	resp, err := http.Get(server.URL + "/metrics")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	want := `messages_processed_total{queue="geofence-test",result="success",service="geofence-service"} 1`
	if !strings.Contains(string(body), want) {
		t.Errorf("/metrics misses %q", want)
	}

	health, err := http.Get(server.URL + "/health")
	if err != nil {
		t.Fatal(err)
	}
	health.Body.Close()
	if health.StatusCode != http.StatusOK {
		t.Errorf("/health = %d", health.StatusCode)
	}
}
