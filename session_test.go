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

func TestHandleControlHostResize(t *testing.T) {
	// arrange
	var gotCols, gotRows uint16
	s := &Session{resizePTY: func(cols, rows uint16) error {
		gotCols, gotRows = cols, rows
		return nil
	}}
	host := &Client{isHost: true}

	// act
	host.handleControl(s, []byte(`{"type":"resize","cols":120,"rows":40}`))

	// assert
	if gotCols != 120 || gotRows != 40 {
		t.Fatalf("want 120x40, got %dx%d", gotCols, gotRows)
	}
}

func TestHandleControlViewerResizeIgnored(t *testing.T) {
	// arrange
	called := false
	s := &Session{resizePTY: func(cols, rows uint16) error {
		called = true
		return nil
	}}
	viewer := &Client{isHost: false}

	// act
	viewer.handleControl(s, []byte(`{"type":"resize","cols":120,"rows":40}`))

	// assert
	if called {
		t.Fatal("viewer resize must be ignored")
	}
}

func TestHandleControlResizeZeroRejected(t *testing.T) {
	// arrange
	called := false
	s := &Session{resizePTY: func(cols, rows uint16) error {
		called = true
		return nil
	}}
	host := &Client{isHost: true}

	// act
	host.handleControl(s, []byte(`{"type":"resize","cols":0,"rows":40}`))

	// assert
	if called {
		t.Fatal("zero cols/rows must be rejected")
	}
}

func TestExitMessage(t *testing.T) {
	// act
	msg := exitMessage()
	var got map[string]any
	json.Unmarshal(msg.data, &got)

	// assert
	if !msg.text {
		t.Fatal("exit frame must be a text message")
	}
	if got["type"] != "exit" {
		t.Fatalf("want type=exit, got %v", got)
	}
}

func TestSessionShutdownNotifiesAndDrops(t *testing.T) {
	// arrange
	removed := false
	s := &Session{clients: map[*Client]bool{}, remove: func() { removed = true }}
	c := &Client{send: make(chan outMsg, 1)}
	s.clients[c] = true

	// act
	s.shutdown()

	// assert
	msg := <-c.send
	var got map[string]any
	json.Unmarshal(msg.data, &got)
	if got["type"] != "exit" {
		t.Fatalf("want exit frame, got %v", got)
	}
	if _, ok := <-c.send; ok {
		t.Fatal("client send channel should be closed after shutdown")
	}
	if len(s.clients) != 0 {
		t.Fatal("clients should be dropped on shutdown")
	}
	if !removed {
		t.Fatal("shutdown should unregister the session")
	}
}

func TestOfferRejectsAfterDone(t *testing.T) {
	// arrange
	s := &Session{register: make(chan *Client), done: make(chan struct{})}
	close(s.done)

	// act
	ok := s.offer(&Client{})

	// assert
	if ok {
		t.Fatal("offer must reject after the session has ended")
	}
}
