package main

import (
	"os"
	"sync"

	"github.com/gorilla/websocket"
)

// session.go is the hub. One Session owns one pty and the set of
// browser clients watching it. Model to copy: the gorilla/websocket
// chat example (hub.go + client.go) — termshare IS that example with
// the pty master file substituted for "messages users type".

// Client is one connected browser.
type Client struct {
	conn *websocket.Conn
	// send is buffered. Broadcasts push here; writePump is the ONLY
	// thing that drains it and writes to conn. This is how you obey
	// gorilla's "one concurrent writer per connection" rule.
	send chan []byte
}

// writePump: the single writer for this client's conn.
//
//	for b := range c.send {
//	    c.conn.WriteMessage(websocket.BinaryMessage, b)  // err -> return
//	}
func (c *Client) writePump() {
	// TODO
}

// readPump: read browser -> pty input. Pass `s` so you can write to
// the pty and unregister on disconnect.
//
//	for { _, b, err := c.conn.ReadMessage(); if err != nil { break }
//	      s.ptyFile.Write(b) }   // skip the Write for read-only viewers
//	defer: s.unregister <- c
func (c *Client) readPump(s *Session) {
	// TODO
}

// Session ties the pty to its clients.
type Session struct {
	ptyFile    *os.File
	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan []byte // bytes read off the pty land here
	mu         sync.Mutex  // only if you guard `clients` directly
	// instead of doing everything through the channels above
}

// NewSession: StartPTY(), build the maps/channels, return *Session.
// Caller then does: go s.run(); go s.readPTY().
func NewSession() (*Session, error) {
	// TODO:
	//   f, _, err := StartPTY()
	//   return &Session{ptyFile: f, clients: map[*Client]bool{},
	//       register: make(chan *Client), unregister: make(chan *Client),
	//       broadcast: make(chan []byte, 256)}, err
	panic("TODO: implement NewSession")
}

// run is the hub loop. select over the three channels:
//
//	case c := <-s.register:    s.clients[c] = true
//	case c := <-s.unregister:  delete + close(c.send) + c.conn.Close()
//	case b := <-s.broadcast:
//	    for c := range s.clients {
//	        select {
//	        case c.send <- b:        // ok
//	        default:                 // slow client: drop it
//	            unregister c
//	        }
//	    }
func (s *Session) run() {
	// TODO
}

// readPTY: the producer. Loop reading the pty, push to broadcast.
//
//	buf := make([]byte, 4096)
//	for {
//	    n, err := s.ptyFile.Read(buf)
//	    if err != nil { /* shell exited: tear the session down */ return }
//	    data := make([]byte, n); copy(data, buf[:n]) // copy! buf is reused
//	    s.broadcast <- data
//	}
func (s *Session) readPTY() {
	// TODO
}
