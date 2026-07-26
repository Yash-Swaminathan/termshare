package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func setupHub(t *testing.T) *Session {
	t.Helper()
	hub = NewHub()
	s := &Session{id: "known123", hostKey: "secretkey"}
	hub.put(s)
	return s
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
