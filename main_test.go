package main

import (
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func setupHub(t *testing.T) *Session {
	t.Helper()
	hub = NewHub()
	s := &Session{id: "known123", hostKey: "secretkey"}
	hub.put(s)
	return s
}

func TestShareURLsJSON(t *testing.T) {
	line, err := shareURLsJSONWithLAN(":9090", "abc123", "hostsecret", "")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(line, "\n") {
		t.Fatalf("want single line, got %q", line)
	}

	var got struct {
		Viewer    string `json:"viewer"`
		Host      string `json:"host"`
		LanViewer string `json:"lanViewer"`
		LanHost   string `json:"lanHost"`
		ID        string `json:"id"`
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
	if got.LanViewer != "" || got.LanHost != "" {
		t.Fatalf("want no LAN fields when lanIP empty, got lanViewer=%q lanHost=%q", got.LanViewer, got.LanHost)
	}
	u, err := url.Parse(got.Host)
	if err != nil {
		t.Fatal(err)
	}
	if u.Query().Get("key") != "hostsecret" {
		t.Fatalf("host key query: want hostsecret, got %q", u.Query().Get("key"))
	}
}

func TestShareURLsJSONWithLAN(t *testing.T) {
	line, err := shareURLsJSONWithLAN(":8080", "sess1", "k", "192.168.1.42")
	if err != nil {
		t.Fatal(err)
	}
	var got struct {
		Viewer    string `json:"viewer"`
		LanViewer string `json:"lanViewer"`
		LanHost   string `json:"lanHost"`
	}
	if err := json.Unmarshal([]byte(line), &got); err != nil {
		t.Fatal(err)
	}
	if got.LanViewer != "http://192.168.1.42:8080/s/sess1" {
		t.Fatalf("lanViewer: got %q", got.LanViewer)
	}
	if got.LanHost != "http://192.168.1.42:8080/s/sess1?key=k" {
		t.Fatalf("lanHost: got %q", got.LanHost)
	}
	if strings.Contains(got.LanViewer, "key=") {
		t.Fatal("lanViewer must not include key")
	}
}

func TestLanIPScore(t *testing.T) {
	if lanIPScore(net.ParseIP("192.168.1.5").To4()) <= lanIPScore(net.ParseIP("10.0.0.5").To4()) {
		t.Fatal("want 192.168 scored higher than 10/8")
	}
	if lanIPScore(net.ParseIP("10.0.0.5").To4()) <= lanIPScore(net.ParseIP("172.17.0.5").To4()) {
		t.Fatal("want 10/8 scored higher than 172.16/12")
	}
}

func TestResolveLANIPOverride(t *testing.T) {
	if got := resolveLANIP("192.168.86.100"); got != "192.168.86.100" {
		t.Fatalf("override: got %q", got)
	}
}

func TestDetectLANIPSkipsLoopbackAlias(t *testing.T) {
	ip := detectLANIP()
	if ip == "10.255.255.254" || ip == "127.0.0.1" {
		t.Fatalf("detectLANIP returned unusable address %q", ip)
	}
}

func TestPreferredViewerURLUsesLAN(t *testing.T) {
	got := preferredViewerURL(":8080", "abc", "192.168.1.42")
	if got != "http://192.168.1.42:8080/s/abc" {
		t.Fatalf("got %q", got)
	}
}

func TestPreferredViewerURLFallsBackToLocalhost(t *testing.T) {
	got := preferredViewerURL(":9090", "abc", "")
	if got != "http://localhost:9090/s/abc" {
		t.Fatalf("got %q", got)
	}
}

func TestShareURLsJSONDefaultPort(t *testing.T) {
	line, err := shareURLsJSONWithLAN("bad-addr", "id1", "k", "")
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
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), "xterm") {
		t.Fatalf("session page should serve the embedded xterm UI, got %q", body)
	}
}

func TestStaticRootServesIndex(t *testing.T) {
	// arrange
	setupHub(t)
	srv := httptest.NewServer(newMux())
	defer srv.Close()

	// act
	resp, err := http.Get(srv.URL + "/index.html")

	// assert
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("embedded index.html: want 200, got %d", resp.StatusCode)
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

// setupPipeSession registers a session whose pty is a pipe, so tests can read
// what the server writes to the shell without a real pty (works on Windows).
func setupPipeSession(t *testing.T) *os.File {
	t.Helper()
	pr, pw, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	hub = NewHub()
	s := &Session{
		id:         "known123",
		hostKey:    "secretkey",
		ptyFile:    pw,
		clients:    map[*Client]bool{},
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan []byte, 256),
		setACL:     make(chan bool),
		done:       make(chan struct{}),
	}
	hub.put(s)
	go s.run()
	t.Cleanup(func() { pr.Close(); pw.Close() })
	return pr
}

func dialHost(t *testing.T, srvURL, path string) *websocket.Conn {
	t.Helper()
	wsURL := "ws" + strings.TrimPrefix(srvURL, "http") + path
	c, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial %s: %v", wsURL, err)
	}
	t.Cleanup(func() { c.Close() })
	return c
}

func TestHostBinaryInputReachesPTY(t *testing.T) {
	// arrange
	pr := setupPipeSession(t)
	srv := httptest.NewServer(newMux())
	defer srv.Close()
	c := dialHost(t, srv.URL, "/s/known123/ws?key=secretkey")

	// act: a text frame is control (dropped here), a binary frame is pty input
	if err := c.WriteMessage(websocket.TextMessage, []byte("echo IGNORED\n")); err != nil {
		t.Fatal(err)
	}
	if err := c.WriteMessage(websocket.BinaryMessage, []byte("echo REAL\n")); err != nil {
		t.Fatal(err)
	}

	// assert: only the binary payload is written to the pty
	got := readWithTimeout(t, pr, len("echo REAL\n"))
	if got != "echo REAL\n" {
		t.Fatalf("pty input: want %q, got %q", "echo REAL\n", got)
	}
}

func readWithTimeout(t *testing.T, r io.Reader, n int) string {
	t.Helper()
	type result struct {
		s   string
		err error
	}
	ch := make(chan result, 1)
	go func() {
		buf := make([]byte, n)
		_, err := io.ReadFull(r, buf)
		ch <- result{string(buf), err}
	}()
	select {
	case res := <-ch:
		if res.err != nil {
			t.Fatalf("read pty: %v", res.err)
		}
		return res.s
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for pty input; typed keystrokes may not be reaching the shell")
		return ""
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
