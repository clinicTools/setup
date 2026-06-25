package ws

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = 30 * time.Second
	outboxCapacity = 256
)

// Handler nimmt WebSocket-Verbindungen entgegen und verwaltet ihre Subscriptions.
type Handler struct {
	registry Registry
	activity *Activity
	upgrader websocket.Upgrader
}

// NewHandler erzeugt einen WS-Handler. dev lockert die Origin-Prüfung für den
// Vite-Dev-Server.
func NewHandler(registry Registry, activity *Activity, dev bool) *Handler {
	return &Handler{
		registry: registry,
		activity: activity,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 4096,
			CheckOrigin:     originChecker(dev),
		},
	}
}

// originChecker erlaubt nur gleichartige Origins (bzw. alles im Dev-Modus).
func originChecker(dev bool) func(*http.Request) bool {
	return func(r *http.Request) bool {
		if dev {
			return true
		}
		origin := r.Header.Get("Origin")
		if origin == "" {
			return true // Nicht-Browser-Clients
		}
		u, err := url.Parse(origin)
		if err != nil {
			return false
		}
		return u.Host == r.Host
	}
}

// ServeHTTP führt den Upgrade durch und betreibt die Verbindung.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	socket, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return // Upgrade-Fehler werden von gorilla bereits beantwortet.
	}

	h.activity.Connect()
	defer h.activity.Disconnect()

	c := &conn{
		socket:   socket,
		registry: h.registry,
		activity: h.activity,
		outbox:   make(chan ServerMessage, outboxCapacity),
		subs:     map[string]context.CancelFunc{},
	}
	c.run(r.Context())
}

// conn kapselt eine einzelne WebSocket-Verbindung samt ihrer Subscriptions.
type conn struct {
	socket   *websocket.Conn
	registry Registry
	activity *Activity
	outbox   chan ServerMessage

	mu   sync.Mutex
	subs map[string]context.CancelFunc
}

// run startet Lese- und Schreibschleife und blockiert bis zum Verbindungsende.
func (c *conn) run(parent context.Context) {
	ctx, cancel := context.WithCancel(parent)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		c.writePump(ctx)
	}()

	c.readPump(ctx, cancel)

	// Verbindung beendet: alle Subscriptions stoppen.
	c.mu.Lock()
	for _, stop := range c.subs {
		stop()
	}
	c.subs = map[string]context.CancelFunc{}
	c.mu.Unlock()

	wg.Wait()
}

// readPump verarbeitet eingehende Steuernachrichten.
func (c *conn) readPump(ctx context.Context, cancel context.CancelFunc) {
	defer cancel()
	c.socket.SetReadLimit(64 * 1024)
	_ = c.socket.SetReadDeadline(time.Now().Add(pongWait))
	c.socket.SetPongHandler(func(string) error {
		_ = c.socket.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, data, err := c.socket.ReadMessage()
		if err != nil {
			return
		}
		c.activity.Touch()

		var msg ClientMessage
		if json.Unmarshal(data, &msg) != nil {
			c.send(ServerMessage{Type: "error", Error: "ungültige Nachricht"})
			continue
		}
		switch msg.Type {
		case "subscribe":
			c.subscribe(ctx, msg)
		case "unsubscribe":
			c.unsubscribe(msg.ID)
		case "ping":
			c.send(ServerMessage{Type: "pong"})
		}
	}
}

// subscribe startet den Collector eines Channels in einer eigenen Goroutine,
// deren Context beim Unsubscribe/Disconnect abgebrochen wird.
func (c *conn) subscribe(parent context.Context, msg ClientMessage) {
	run, ok := c.registry[msg.Channel]
	if !ok {
		c.send(ServerMessage{Type: "error", ID: msg.ID, Error: "unbekannter Channel: " + msg.Channel})
		return
	}

	c.mu.Lock()
	if _, exists := c.subs[msg.ID]; exists {
		c.mu.Unlock()
		c.send(ServerMessage{Type: "error", ID: msg.ID, Error: "Subscription-ID bereits aktiv"})
		return
	}
	subCtx, cancel := context.WithCancel(parent)
	c.subs[msg.ID] = cancel
	c.mu.Unlock()

	c.send(ServerMessage{Type: "ready", ID: msg.ID, Channel: msg.Channel})

	emit := func(payload any) {
		c.send(ServerMessage{Type: "message", ID: msg.ID, Channel: msg.Channel, Payload: payload})
	}

	go func() {
		defer c.unsubscribe(msg.ID)
		if err := run(subCtx, msg.Params, emit); err != nil && subCtx.Err() == nil {
			c.send(ServerMessage{Type: "error", ID: msg.ID, Channel: msg.Channel, Error: err.Error()})
		}
	}()
}

// unsubscribe stoppt den Collector einer Subscription (idempotent).
func (c *conn) unsubscribe(id string) {
	c.mu.Lock()
	cancel, ok := c.subs[id]
	if ok {
		delete(c.subs, id)
	}
	c.mu.Unlock()
	if ok {
		cancel()
		c.send(ServerMessage{Type: "closed", ID: id})
	}
}

// send stellt eine Nachricht in die Outbox; bei voller Outbox wird verworfen,
// damit ein langsamer Client die Collector-Goroutinen nicht blockiert.
func (c *conn) send(msg ServerMessage) {
	select {
	case c.outbox <- msg:
	default:
		log.Printf("ws: Outbox voll, Nachricht (%s/%s) verworfen", msg.Type, msg.Channel)
	}
}

// writePump serialisiert alle ausgehenden Nachrichten (gorilla erlaubt nur
// einen Schreiber) und sendet periodische Pings.
func (c *conn) writePump(ctx context.Context) {
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()
	defer func() { _ = c.socket.Close() }()

	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-c.outbox:
			_ = c.socket.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.socket.WriteJSON(msg); err != nil {
				return
			}
		case <-ticker.C:
			_ = c.socket.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.socket.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
