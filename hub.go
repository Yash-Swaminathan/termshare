package main

import "sync"

// Hub is the registry of live sessions keyed by public session id.
type Hub struct {
	mu       sync.Mutex
	sessions map[string]*Session
}

func NewHub() *Hub {
	return &Hub{sessions: map[string]*Session{}}
}

func (h *Hub) Create(hostKey string) (*Session, error) {
	id := randomKey()
	s, err := NewSession(id, hostKey)
	if err != nil {
		return nil, err
	}
	s.remove = func() { h.delete(s.id) }
	h.put(s)
	return s, nil
}

// put is separate from Create so tests can register a session without a pty.
func (h *Hub) put(s *Session) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.sessions[s.id] = s
}

func (h *Hub) Get(id string) (*Session, bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	s, ok := h.sessions[id]
	return s, ok
}

func (h *Hub) delete(id string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.sessions, id)
}
