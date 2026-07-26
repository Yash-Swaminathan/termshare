package main

import (
	"encoding/json"
	"os"
	"sync/atomic"

	"github.com/gorilla/websocket"
)

// outMsg is one queued frame: binary for pty output, text for JSON control.
type outMsg struct {
	text bool
	data []byte
}

// Client is one connected browser. canWrite is read by readPump and written
// by the hub, so it is atomic.
type Client struct {
	conn     *websocket.Conn
	send     chan outMsg
	isHost   bool
	canWrite atomic.Bool
}

// trySend queues a frame without blocking; false means the client is too slow.
func (c *Client) trySend(m outMsg) bool {
	select {
	case c.send <- m:
		return true
	default:
		return false
	}
}

func (c *Client) roleMessage(viewersCanWrite bool) outMsg {
	role := "viewer"
	if c.isHost {
		role = "host"
	}
	data, _ := json.Marshal(struct {
		Type            string `json:"type"`
		Role            string `json:"role"`
		CanWrite        bool   `json:"canWrite"`
		ViewersCanWrite bool   `json:"viewersCanWrite"`
	}{"role", role, c.canWrite.Load(), viewersCanWrite})
	return outMsg{text: true, data: data}
}

func (c *Client) writePump() {
	for m := range c.send {
		mt := websocket.BinaryMessage
		if m.text {
			mt = websocket.TextMessage
		}
		c.conn.WriteMessage(mt, m.data)
	}
}

func (c *Client) readPump(s *Session) {
	defer func() { s.unregister <- c }()
	for {
		mt, b, err := c.conn.ReadMessage()
		if err != nil {
			break
		}
		if mt == websocket.TextMessage {
			c.handleControl(s, b)
			continue
		}
		if c.canWrite.Load() {
			s.ptyFile.Write(b)
		}
	}
}

func (c *Client) handleControl(s *Session, b []byte) {
	var m struct {
		Type         string `json:"type"`
		ViewersWrite bool   `json:"viewersWrite"`
	}
	if err := json.Unmarshal(b, &m); err != nil {
		return
	}
	if m.Type == "set_acl" && c.isHost {
		s.setACL <- m.ViewersWrite
	}
}

// Session ties the pty to its clients. hostKey grants write via ?key=; all
// role state is mutated only in run().
type Session struct {
	ptyFile    *os.File
	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan []byte
	setACL     chan bool

	hostKey         string
	viewersCanWrite bool
}

func NewSession(hostKey string) (*Session, error) {
	f, _, err := StartPTY()
	if err != nil {
		return nil, err
	}
	return &Session{
		ptyFile:    f,
		clients:    map[*Client]bool{},
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan []byte, 256),
		setACL:     make(chan bool),
		hostKey:    hostKey,
	}, nil
}

func (s *Session) run() {
	for {
		select {
		case c := <-s.register:
			s.clients[c] = true
			if !c.isHost {
				c.canWrite.Store(s.viewersCanWrite)
			}
			c.trySend(c.roleMessage(s.viewersCanWrite))
		case c := <-s.unregister:
			s.drop(c)
		case vw := <-s.setACL:
			s.viewersCanWrite = vw
			for c := range s.clients {
				if !c.isHost {
					c.canWrite.Store(vw)
				}
				c.trySend(c.roleMessage(vw))
			}
		case b := <-s.broadcast:
			for c := range s.clients {
				// Drop slow clients inline; unregister would deadlock here.
				if !c.trySend(outMsg{data: b}) {
					s.drop(c)
				}
			}
		}
	}
}

// drop removes a client exactly once (safe to call twice).
func (s *Session) drop(c *Client) {
	if _, ok := s.clients[c]; ok {
		delete(s.clients, c)
		close(c.send)
		c.conn.Close()
	}
}

func (s *Session) readPTY() {
	buf := make([]byte, 4096)
	for {
		n, err := s.ptyFile.Read(buf)
		if err != nil {
			return
		}
		data := make([]byte, n)
		copy(data, buf[:n])
		s.broadcast <- data
	}
}
