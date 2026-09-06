package websocket

import (
    "encoding/json"
    "log"
    "sync"

    "github.com/google/uuid"
)

// Message estructura de mensaje WebSocket
type Message struct {
    Type    string      `json:"type"`
    Data    interface{} `json:"data"`
    UserID  string      `json:"user_id,omitempty"`
    RideID  string      `json:"ride_id,omitempty"`
}

// Client representa un cliente WebSocket
type Client struct {
    ID     uuid.UUID
    UserID uuid.UUID
    Hub    *Hub
    Send   chan []byte
}

// Hub mantiene los clientes activos y transmite mensajes
type Hub struct {
    mu       sync.RWMutex
    Clients  map[uuid.UUID]*Client
    Register chan *Client
    Unregister chan *Client
    Broadcast  chan []byte
    UserRooms map[uuid.UUID]map[uuid.UUID]bool // userID -> set de clientIDs
    RideRooms map[uuid.UUID]map[uuid.UUID]bool // rideID -> set de clientIDs
}

// NewHub crea un nuevo Hub
func NewHub() *Hub {
    return &Hub{
        Clients:    make(map[uuid.UUID]*Client),
        Register:   make(chan *Client),
        Unregister: make(chan *Client),
        Broadcast:  make(chan []byte),
        UserRooms:  make(map[uuid.UUID]map[uuid.UUID]bool),
        RideRooms:  make(map[uuid.UUID]map[uuid.UUID]bool),
    }
}

// Run inicia el Hub
func (h *Hub) Run() {
    for {
        select {
        case client := <-h.Register:
            h.mu.Lock()
            h.Clients[client.ID] = client
            
            // Agregar a room de usuario
            if h.UserRooms[client.UserID] == nil {
                h.UserRooms[client.UserID] = make(map[uuid.UUID]bool)
            }
            h.UserRooms[client.UserID][client.ID] = true
            h.mu.Unlock()
            
            log.Printf("✅ Cliente conectado: %s (User: %s)", client.ID, client.UserID)

        case client := <-h.Unregister:
            h.mu.Lock()
            if _, ok := h.Clients[client.ID]; ok {
                delete(h.Clients, client.ID)
                close(client.Send)
                
                // Eliminar de room de usuario
                if h.UserRooms[client.UserID] != nil {
                    delete(h.UserRooms[client.UserID], client.ID)
                    if len(h.UserRooms[client.UserID]) == 0 {
                        delete(h.UserRooms, client.UserID)
                    }
                }
            }
            h.mu.Unlock()
            
            log.Printf("🔌 Cliente desconectado: %s", client.ID)

        case message := <-h.Broadcast:
            h.mu.RLock()
            for _, client := range h.Clients {
                select {
                case client.Send <- message:
                default:
                    close(client.Send)
                    delete(h.Clients, client.ID)
                }
            }
            h.mu.RUnlock()
        }
    }
}

// SendToUser envía mensaje a todos los clientes de un usuario
func (h *Hub) SendToUser(userID uuid.UUID, message Message) {
    h.mu.RLock()
    defer h.mu.RUnlock()
    
    if clients, ok := h.UserRooms[userID]; ok {
        data, _ := json.Marshal(message)
        for clientID := range clients {
            if client, exists := h.Clients[clientID]; exists {
                select {
                case client.Send <- data:
                default:
                }
            }
        }
    }
}

// SendToRide envía mensaje a todos los clientes de un ride
func (h *Hub) SendToRide(rideID uuid.UUID, message Message) {
    h.mu.RLock()
    defer h.mu.RUnlock()
    
    if clients, ok := h.RideRooms[rideID]; ok {
        data, _ := json.Marshal(message)
        for clientID := range clients {
            if client, exists := h.Clients[clientID]; exists {
                select {
                case client.Send <- data:
                default:
                }
            }
        }
    }
}

// JoinRideRoom agrega un cliente a la room de un ride
func (h *Hub) JoinRideRoom(client *Client, rideID uuid.UUID) {
    h.mu.Lock()
    defer h.mu.Unlock()
    
    if h.RideRooms[rideID] == nil {
        h.RideRooms[rideID] = make(map[uuid.UUID]bool)
    }
    h.RideRooms[rideID][client.ID] = true
}

// LeaveRideRoom elimina un cliente de la room de un ride
func (h *Hub) LeaveRideRoom(client *Client, rideID uuid.UUID) {
    h.mu.Lock()
    defer h.mu.Unlock()
    
    if h.RideRooms[rideID] != nil {
        delete(h.RideRooms[rideID], client.ID)
        if len(h.RideRooms[rideID]) == 0 {
            delete(h.RideRooms, rideID)
        }
    }
}
