package main

import (
	"encoding/json"
	"testing"
)

func TestRoleMessageHost(t *testing.T) {
	// arrange
	c := &Client{isHost: true}
	c.canWrite.Store(true)

	// act
	msg := c.roleMessage(false)
	var got map[string]any
	json.Unmarshal(msg.data, &got)

	// assert
	if !msg.text {
		t.Fatal("role frame must be a text message")
	}
	if got["role"] != "host" || got["canWrite"] != true {
		t.Fatalf("want host/canWrite=true, got %v", got)
	}
}

func TestRoleMessageViewer(t *testing.T) {
	// arrange
	c := &Client{isHost: false}

	// act
	msg := c.roleMessage(false)
	var got map[string]any
	json.Unmarshal(msg.data, &got)

	// assert
	if got["role"] != "viewer" || got["canWrite"] != false {
		t.Fatalf("want viewer/canWrite=false, got %v", got)
	}
}

func TestTrySendDropsWhenFull(t *testing.T) {
	// arrange
	c := &Client{send: make(chan outMsg, 1)}
	c.send <- outMsg{}

	// act
	ok := c.trySend(outMsg{})

	// assert
	if ok {
		t.Fatal("trySend should fail when the buffer is full")
	}
}

func TestHandleControlHostSetsACL(t *testing.T) {
	// arrange
	s := &Session{setACL: make(chan bool, 1)}
	host := &Client{isHost: true}

	// act
	host.handleControl(s, []byte(`{"type":"set_acl","viewersWrite":true}`))

	// assert
	if got := <-s.setACL; got != true {
		t.Fatalf("host set_acl should push true, got %v", got)
	}
}

func TestHandleControlViewerIgnored(t *testing.T) {
	// arrange
	s := &Session{setACL: make(chan bool, 1)}
	viewer := &Client{isHost: false}

	// act
	viewer.handleControl(s, []byte(`{"type":"set_acl","viewersWrite":true}`))

	// assert
	select {
	case <-s.setACL:
		t.Fatal("viewer set_acl must be ignored")
	default:
	}
}
