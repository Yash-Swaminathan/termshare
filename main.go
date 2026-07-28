package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/websocket"
)

var (
	addr      = flag.String("addr", ":8080", "http service address")
	hostKey   = flag.String("host-key", "", "key that grants host (write) access; random if empty")
	lanIPFlag = flag.String("lan-ip", "", "LAN IP for share URLs (overrides auto-detect; useful under WSL)")
	printJSON = flag.Bool("print-json", false, "print one machine-readable share URL line to stdout")
)

func main() {
	flag.Parse()

	key := *hostKey
	if key == "" {
		key = randomKey()
	}

	hub = NewHub()
	s, err := hub.Create(key)
	if err != nil {
		log.Fatal("Create session:", err)
	}
	go s.run()
	go s.readPTY()
	go s.waitShell()

	log.Printf("termshare listening on %s", *addr)
	lanIP := resolveLANIP(*lanIPFlag)
	logShareURLs(*addr, s.id, key, lanIP)
	if *printJSON {
		line, err := shareURLsJSONWithLAN(*addr, s.id, key, lanIP)
		if err != nil {
			log.Fatal("print-json:", err)
		}
		fmt.Fprintln(os.Stdout, line)
	}
	if err := http.ListenAndServe(*addr, newMux()); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

// newMux requires hub to be initialized before the handler serves requests.
func newMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.FS(staticRoot())))
	mux.HandleFunc("/s/{id}", serveSessionPage)
	mux.HandleFunc("/s/{id}/ws", serveWS)
	return mux
}

// staticRoot returns the embedded UI rooted at static/ so paths are served as
// "/index.html" rather than "/static/index.html".
func staticRoot() fs.FS {
	sub, err := fs.Sub(staticFS, "static")
	if err != nil {
		log.Fatal("static embed:", err)
	}
	return sub
}

func logShareURLs(addr, id, key, lanIP string) {
	port := portOf(addr)
	log.Printf("session  %s", id)
	printPair := func(label, host string) {
		base := "http://" + host + ":" + port + "/s/" + id
		log.Printf("%s viewer  %s", label, base)
		log.Printf("%s host    %s?key=%s", label, base, key)
	}
	printPair("local", "localhost")
	if lanIP != "" {
		printPair("lan  ", lanIP)
	}
}

// shareURLsJSON returns one line of machine-readable share URLs for the
// VS Code extension (and similar tools) to parse without scraping stderr.
// When a LAN IP is available, lanViewer/lanHost are included for same-network sharing.
func shareURLsJSON(addr, id, key string) (string, error) {
	return shareURLsJSONWithLAN(addr, id, key, resolveLANIP(""))
}

func shareURLsJSONWithLAN(addr, id, key, lanIP string) (string, error) {
	port := portOf(addr)
	base := "http://localhost:" + port + "/s/" + id
	payload := struct {
		Viewer    string `json:"viewer"`
		Host      string `json:"host"`
		LanViewer string `json:"lanViewer,omitempty"`
		LanHost   string `json:"lanHost,omitempty"`
		ID        string `json:"id"`
	}{
		Viewer: base,
		Host:   base + "?key=" + key,
		ID:     id,
	}
	if lanIP != "" {
		lanBase := "http://" + lanIP + ":" + port + "/s/" + id
		payload.LanViewer = lanBase
		payload.LanHost = lanBase + "?key=" + key
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func portOf(addr string) string {
	if _, port, err := net.SplitHostPort(addr); err == nil && port != "" {
		return port
	}
	return "8080"
}

func resolveLANIP(override string) string {
	if override != "" {
		return override
	}
	return detectLANIP()
}

// detectLANIP picks a usable IPv4 for same-Wi-Fi share links.
// Skips loopback interfaces and link-local addresses so WSL aliases like
// 10.255.255.254 on lo are not advertised.
func detectLANIP() string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return ""
	}
	var best string
	bestScore := 0
	for _, iface := range ifaces {
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, a := range addrs {
			ipNet, ok := a.(*net.IPNet)
			if !ok {
				continue
			}
			ip4 := ipNet.IP.To4()
			if ip4 == nil || ip4.IsLoopback() || ip4.IsLinkLocalUnicast() {
				continue
			}
			score := lanIPScore(ip4)
			if score > bestScore {
				bestScore = score
				best = ip4.String()
			}
		}
	}
	return best
}

// lanIPScore ranks candidates: prefer typical home/office Wi-Fi (192.168/16),
// then other RFC1918, then anything else unicast.
func lanIPScore(ip net.IP) int {
	if ip[0] == 192 && ip[1] == 168 {
		return 30
	}
	if ip[0] == 10 {
		return 20
	}
	if ip[0] == 172 && ip[1] >= 16 && ip[1] <= 31 {
		// WSL/Hyper-V often lands here; usable on Linux/macOS LAN, weaker on WSL.
		return 10
	}
	return 5
}

func randomKey() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		log.Fatal("randomKey:", err)
	}
	return hex.EncodeToString(b)
}

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	hub *Hub
)

func serveSessionPage(w http.ResponseWriter, req *http.Request) {
	id := strings.TrimSpace(req.PathValue("id"))
	if _, ok := hub.Get(id); !ok {
		http.NotFound(w, req)
		return
	}
	http.ServeFileFS(w, req, staticRoot(), "index.html")
}

func serveWS(w http.ResponseWriter, req *http.Request) {
	id := strings.TrimSpace(req.PathValue("id"))
	s, ok := hub.Get(id)
	if !ok {
		http.NotFound(w, req)
		return
	}
	conn, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		return
	}
	isHost := s.hostKey != "" && req.URL.Query().Get("key") == s.hostKey
	client := &Client{conn: conn, send: make(chan outMsg, 256), isHost: isHost}
	client.canWrite.Store(isHost)
	if !s.offer(client) {
		conn.Close()
		return
	}
	go client.writePump()
	go client.readPump(s)
}
