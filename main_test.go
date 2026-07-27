package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func setupHub(t *testing.T) *Session {
	t.Helper()
	hub = NewHub()
	s := &Session{id: "known123", hostKey: "secretkey"}
	hub.put(s)
	return s
}

func TestShareURLsJSON(t *testing.T) {
	line, err := shareURLsJSON(":9090", "abc123", "hostsecret")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(line, "\n") {
		t.Fatalf("want single line, got %q", line)
	}

	var got struct {
		Viewer string `json:"viewer"`
		Host   string `json:"host"`
		ID     string `json:"id"`
	}
	if err := json.Unmarshal([]byte(line), &got); err != nil {
		t.Fatalf("unmarshal: %v\nline=%q", err, line)
	}
	if got.ID != "abc123" {
		t.Fatalf("id: want abc123, got %q", got.ID)
	}
	if got.Viewer != "http://localhost:9090/s/abc123" {
		t.Fatalf("viewer: want local path without key, got %q", got.Viewer)
	}
	if strings.Contains(got.Viewer, "key=") {
		t.Fatalf("viewer must not include key, got %q", got.Viewer)
	}
	if got.Host != "http://localhost:9090/s/abc123?key=hostsecret" {
		t.Fatalf("host: want key query, got %q", got.Host)
	}
	u, err := url.Parse(got.Host)
	if err != nil {
		t.Fatal(err)
	}
	if u.Query().Get("key") != "hostsecret" {
		t.Fatalf("host key query: want hostsecret, got %q", u.Query().Get("key"))
	}
}

func TestShareURLsJSONDefaultPort(t *testing.T) {
	line, err := shareURLsJSON("bad-addr", "id1", "k")
	if err != nil {
		t.Fatal(err)
	}
	var got struct {
		Viewer string `json:"viewer"`
	}
	if err := json.Unmarshal([]byte(line), &got); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got.Viewer, ":8080/") {
		t.Fatalf("want fallback port 8080, got %q", got.Viewer)
	}
}

func TestSessionPageKnownID(t *testing.T) {
	// arrange
	setupHub(t)
	srv := httptest.NewServer(newMux())
	defer srv.Close()

	// act
	resp, err := http.Get(srv.URL + "/s/known123")

	// assert
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("known session page: want 200, got %d", resp.StatusCode)
	}
}

func TestSessionPageUnknownID404(t *testing.T) {
	// arrange
	setupHub(t)
	srv := httptest.NewServer(newMux())
	defer srv.Close()

	// act
	resp, err := http.Get(srv.URL + "/s/deadbeef")

	// assert
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("unknown session page: want 404, got %d", resp.StatusCode)
	}
}

func TestWSUnknownID404(t *testing.T) {
	// arrange
	setupHub(t)
	srv := httptest.NewServer(newMux())
	defer srv.Close()

	// act
	resp, err := http.Get(srv.URL + "/s/deadbeef/ws")

	// assert
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("unknown ws: want 404, got %d", resp.StatusCode)
	}
}

func TestWSKnownIDRejectsNonUpgrade(t *testing.T) {
	// arrange
	setupHub(t)
	srv := httptest.NewServer(newMux())
	defer srv.Close()

	// act
	resp, err := http.Get(srv.URL + "/s/known123/ws")

	// assert
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("known ws without upgrade: want 400, got %d", resp.StatusCode)
	}
}
