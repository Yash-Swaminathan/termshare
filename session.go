package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"sync"
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

// exitMessage tells clients the shared shell has ended.
func exitMessage() outMsg {
	data, _ := json.Marshal(struct {
		Type string `json:"type"`
	}{"exit"})
	return outMsg{text: true, data: data}
}

// countMessage reports how many viewers are watching.
func countMessage(n int) outMsg {
	data, _ := json.Marshal(struct {
		Type    string `json:"type"`
		Viewers int    `json:"viewers"`
	}{"count", n})
	return outMsg{text: true, data: data}
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
		Cols         uint16 `json:"cols"`
		Rows         uint16 `json:"rows"`
	}
	if err := json.Unmarshal(b, &m); err != nil {
		return
	}
	switch m.Type {
	case "set_acl":
		if c.isHost {
			s.setACL <- m.ViewersWrite
		}
	case "resize":
		if c.isHost {
			s.resize(m.Cols, m.Rows)
		}
	}
}

// Session ties the pty to its clients. hostKey grants write via ?key=; all
// role state is mutated only in run().
type Session struct {
	id         string
	ptyFile    *os.File
	cmd        *exec.Cmd
	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan []byte
	setACL     chan bool
	done       chan struct{}
	doneOnce   sync.Once

	hostKey         string
	viewersCanWrite bool

	// remove unregisters this session from the hub during shutdown.
	remove func()

	// resizePTY applies a new pty size; swappable in tests to avoid a real pty.
	resizePTY func(cols, rows uint16) error
}

// resize validates and applies a host-requested pty size.
func (s *Session) resize(cols, rows uint16) {
	const maxDim = 1000
	if cols == 0 || rows == 0 || cols > maxDim || rows > maxDim {
		return
	}
	s.resizePTY(cols, rows)
}

func NewSession(id, hostKey string) (*Session, error) {
	f, cmd, err := StartPTY()
	if err != nil {
		return nil, err
	}
	s := &Session{
		id:         id,
		ptyFile:    f,
		cmd:        cmd,
		clients:    map[*Client]bool{},
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan []byte, 256),
		setACL:     make(chan bool),
		done:       make(chan struct{}),
		hostKey:    hostKey,
	}
	s.resizePTY = func(cols, rows uint16) error {
		return SetPTYSize(s.ptyFile, cols, rows)
	}
	return s, nil
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
			s.broadcastCount()
		case c := <-s.unregister:
			s.drop(c)
			s.broadcastCount()
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
		case <-s.done:
			s.shutdown()
			return
		}
	}
}

// offer hands a client to run(), or reports false if the session has ended.
func (s *Session) offer(c *Client) bool {
	select {
	case s.register <- c:
		return true
	case <-s.done:
		return false
	}
}

// viewerCount is the number of connected non-host clients.
func (s *Session) viewerCount() int {
	n := 0
	for c := range s.clients {
		if !c.isHost {
			n++
		}
	}
	return n
}

// broadcastCount pushes the current viewer count to every client.
func (s *Session) broadcastCount() {
	msg := countMessage(s.viewerCount())
	for c := range s.clients {
		c.trySend(msg)
	}
}

// shutdown notifies clients the shell exited, drops them, and unregisters.
func (s *Session) shutdown() {
	msg := exitMessage()
	for c := range s.clients {
		c.trySend(msg)
	}
	for c := range s.clients {
		s.drop(c)
	}
	if s.remove != nil {
		s.remove()
	}
}

// drop removes a client exactly once (safe to call twice).
func (s *Session) drop(c *Client) {
	if _, ok := s.clients[c]; ok {
		delete(s.clients, c)
		close(c.send)
		if c.conn != nil {
			c.conn.Close()
		}
	}
}

func (s *Session) readPTY() {
	buf := make([]byte, 4096)
	for {
		n, err := s.ptyFile.Read(buf)
		if err != nil {
			s.signalDone()
			return
		}
		data := make([]byte, n)
		copy(data, buf[:n])
		s.broadcast <- data
	}
}

// waitShell reaps the shell process, then closes the pty to unblock readPTY.
func (s *Session) waitShell() {
	s.cmd.Wait()
	s.ptyFile.Close()
	s.signalDone()
}

// signalDone closes done exactly once so run() shuts the session down.
func (s *Session) signalDone() {
	s.doneOnce.Do(func() { close(s.done) })
}
