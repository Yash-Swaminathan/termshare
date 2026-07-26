package main

import "testing"

func TestHubGetReturnsRegisteredSession(t *testing.T) {
	// arrange
	h := NewHub()
	s := &Session{id: "abc123"}
	h.put(s)

	// act
	got, ok := h.Get("abc123")

	// assert
	if !ok {
		t.Fatal("Get should find a registered session")
	}
	if got != s {
		t.Fatalf("Get returned a different session: %+v", got)
	}
}

func TestHubGetUnknownIDMisses(t *testing.T) {
	// arrange
	h := NewHub()

	// act
	_, ok := h.Get("nope")

	// assert
	if ok {
		t.Fatal("Get should miss for an unknown id")
	}
}

func TestHubDeleteRemovesSession(t *testing.T) {
	// arrange
	h := NewHub()
	h.put(&Session{id: "abc123"})

	// act
	h.delete("abc123")
	_, ok := h.Get("abc123")

	// assert
	if ok {
		t.Fatal("delete should remove the session")
	}
}
