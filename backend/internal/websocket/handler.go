package websocket

import (
    "encoding/json"
    "log"
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "github.com/gorilla/websocket"

)

var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
    CheckOrigin: func(r *http.Request) bool {
        return true // Permitir todas las conexiones en desarrollo
    },
}

// WebSocketHandler maneja conexiones WebSocket
type WebSocketHandler struct {
    Hub *Hub
}

func NewWebSocketHandler(hub *Hub) *WebSocketHandler {
    return &WebSocketHandler{Hub: hub}
}

// HandleConnection maneja la conexión WebSocket
func (h *WebSocketHandler) HandleConnection(c *gin.Context) {
    // Obtener userID del contexto
    userIDInterface, exists := c.Get("user_id")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "No autorizado"})
        return
    }
    userID := userIDInterface.(uuid.UUID)

    // Upgrade a WebSocket
    conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
    if err != nil {
        log.Printf("Error al actualizar a WebSocket: %v", err)
        return
    }

    // Crear cliente
    client := &Client{
        ID:     uuid.New(),
        UserID: userID,
        Hub:    h.Hub,
        Send:   make(chan []byte, 256),
    }

    // Registrar cliente
    h.Hub.Register <- client

    // Goroutines para leer y escribir
    go h.writePump(client, conn)
    go h.readPump(client, conn)
}

// readPump lee mensajes del cliente
func (h *WebSocketHandler) readPump(client *Client, conn *websocket.Conn) {
    defer func() {
        h.Hub.Unregister <- client
        conn.Close()
    }()

    for {
        _, message, err := conn.ReadMessage()
        if err != nil {
            if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
                log.Printf("Error de lectura: %v", err)
            }
            break
        }

        // Procesar mensaje
        var msg Message
        if err := json.Unmarshal(message, &msg); err != nil {
            continue
        }

        // Manejar diferentes tipos de mensajes
        switch msg.Type {
        case "join_ride":
            if rideIDStr, ok := msg.Data.(string); ok {
                if rideID, err := uuid.Parse(rideIDStr); err == nil {
                    h.Hub.JoinRideRoom(client, rideID)
                }
            }
        case "leave_ride":
            if rideIDStr, ok := msg.Data.(string); ok {
                if rideID, err := uuid.Parse(rideIDStr); err == nil {
                    h.Hub.LeaveRideRoom(client, rideID)
                }
            }
        case "ping":
            client.Send <- []byte(`{"type":"pong"}`)
        }
    }
}

// writePump escribe mensajes al cliente
func (h *WebSocketHandler) writePump(client *Client, conn *websocket.Conn) {
    defer func() {
        conn.Close()
    }()

    for {
        select {
        case message, ok := <-client.Send:
            if !ok {
                conn.WriteMessage(websocket.CloseMessage, []byte{})
                return
            }

            if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
                return
            }
        }
    }
}

// SendNotification envía notificación a un usuario
func (h *WebSocketHandler) SendNotification(userID uuid.UUID, notification interface{}) {
    msg := Message{
        Type: "notification",
        Data: notification,
    }
    h.Hub.SendToUser(userID, msg)
}

// SendRideUpdate envía actualización de ride
func (h *WebSocketHandler) SendRideUpdate(rideID uuid.UUID, update interface{}) {
    msg := Message{
        Type:   "ride_update",
        Data:   update,
        RideID: rideID.String(),
    }
    h.Hub.SendToRide(rideID, msg)
}

// SendDriverLocation envía ubicación de conductor
func (h *WebSocketHandler) SendDriverLocation(rideID uuid.UUID, lat, lng float64) {
    msg := Message{
        Type: "driver_location",
        Data: gin.H{
            "lat": lat,
            "lng": lng,
        },
        RideID: rideID.String(),
    }
    h.Hub.SendToRide(rideID, msg)
}
