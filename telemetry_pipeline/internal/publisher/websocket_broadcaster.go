/*
WebSocket broadcaster for publishing filtered state estimates.
Pushes clean, fused state at 60Hz to connected clients.
*/

package publisher

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	estimation "edl-telemetry-reconstructor/internal/estimation"

	"github.com/gorilla/websocket"
)

// StateUpdate represents the filtered state sent to clients
type StateUpdate struct {
	Timestamp   time.Time   `json:"timestamp"`
	Position    [3]float64  `json:"position"`
	Velocity    [3]float64  `json:"velocity"`
	Quaternion  [4]float64  `json:"quaternion"`
	Uncertainty [10]float64 `json:"uncertainty"`
}

// WebSocketBroadcaster broadcasts state estimates to connected clients
type WebSocketBroadcaster struct {
	clients     map[*websocket.Conn]bool
	broadcast   chan *StateUpdate
	register    chan *websocket.Conn
	unregister  chan *websocket.Conn
	mu          sync.RWMutex
	stopChan    chan bool
	wg          sync.WaitGroup
	publishRate time.Duration
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for development
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// NewWebSocketBroadcaster creates a new broadcaster
func NewWebSocketBroadcaster(publishRateHz float64) *WebSocketBroadcaster {
	return &WebSocketBroadcaster{
		clients:     make(map[*websocket.Conn]bool),
		broadcast:   make(chan *StateUpdate, 10),
		register:    make(chan *websocket.Conn),
		unregister:  make(chan *websocket.Conn),
		stopChan:    make(chan bool),
		publishRate: time.Duration(float64(time.Second) / publishRateHz),
	}
}

// Start begins the broadcaster
func (wb *WebSocketBroadcaster) Start() {
	wb.wg.Add(1)
	go wb.run()
	log.Println("WebSocket broadcaster started")
}

// run processes registration, unregistration, and broadcasts
func (wb *WebSocketBroadcaster) run() {
	defer wb.wg.Done()

	for {
		select {
		case client := <-wb.register:
			wb.mu.Lock()
			wb.clients[client] = true
			wb.mu.Unlock()
			log.Printf("WebSocket client connected, total: %d", len(wb.clients))

		case client := <-wb.unregister:
			wb.mu.Lock()
			if ok := wb.clients[client]; ok {
				delete(wb.clients, client)
				client.Close()
			}
			wb.mu.Unlock()
			log.Printf("WebSocket client disconnected, total: %d", len(wb.clients))

		case update := <-wb.broadcast:
			wb.mu.RLock()
			for client := range wb.clients {
				client.SetWriteDeadline(time.Now().Add(10 * time.Second))

				if err := client.WriteJSON(update); err != nil {
					wb.mu.RUnlock()
					wb.unregister <- client
					wb.mu.RLock()
				}
			}
			wb.mu.RUnlock()

		case <-wb.stopChan:
			wb.mu.Lock()
			for client := range wb.clients {
				client.Close()
			}
			wb.mu.Unlock()
			return
		}
	}
}

// PublishState broadcasts the current state
func (wb *WebSocketBroadcaster) PublishState(state estimation.State, uncertainty [10]float64) {
	update := &StateUpdate{
		Timestamp:   state.Timestamp,
		Position:    state.Position,
		Velocity:    state.Velocity,
		Quaternion:  state.Quaternion,
		Uncertainty: uncertainty,
	}

	select {
	case wb.broadcast <- update:
	default:
		// Buffer full, drop update
	}
}

// HandleWebSocket handles incoming WebSocket connections
func (wb *WebSocketBroadcaster) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	wb.register <- conn

	// Keep connection alive, handle incoming messages (if any)
	go func() {
		defer func() {
			wb.unregister <- conn
		}()

		for {
			var msg json.RawMessage
			if err := conn.ReadJSON(&msg); err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("WebSocket error: %v", err)
				}
				return
			}
		}
	}()
}

// Stop stops the broadcaster
func (wb *WebSocketBroadcaster) Stop() {
	wb.stopChan <- true
	wb.wg.Wait()
}

// GetClientCount returns the number of connected clients
func (wb *WebSocketBroadcaster) GetClientCount() int {
	wb.mu.RLock()
	defer wb.mu.RUnlock()
	return len(wb.clients)
}
