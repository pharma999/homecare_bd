package handlers

import (
	"context"
	"log"
	"strings"
	"time"

	"home_care_backend/internal/utils"
	ws "home_care_backend/internal/websocket"

	"github.com/gin-gonic/gin"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

// ConnectWS upgrades an HTTP request to a WebSocket connection.
//
//	GET /api/ws?token=<jwt>&rooms=emergency:abc,admin:emergencies
//
// After connecting the client sends JSON frames:
//
//	{"type":"subscribe","room":"emergency:abc123"}
//	{"type":"ping"}
func ConnectWS(c *gin.Context) {
	// Validate token via query param (headers can't be set by browser WS API)
	token := c.Query("token")
	if token == "" {
		utils.UnauthorizedResponse(c, "token required")
		return
	}
	claims, err := utils.ValidateToken(token)
	if err != nil {
		utils.UnauthorizedResponse(c, "invalid token")
		return
	}

	conn, err := websocket.Accept(c.Writer, c.Request, &websocket.AcceptOptions{
		InsecureSkipVerify: true, // CORS handled by middleware
	})
	if err != nil {
		log.Printf("[WS] upgrade error: %v", err)
		return
	}
	conn.SetReadLimit(4096)

	cl := ws.Global.Register(conn)
	defer ws.Global.Unregister(cl)

	userID := claims.UserID
	userRole := claims.Role
	log.Printf("[WS] connected user=%s role=%s", userID, userRole)

	// Auto-subscribe to user-specific room so they get personal notifications
	ws.Global.Subscribe(cl, "user:"+userID)

	// Admins and super-admins auto-join the emergency broadcast room
	if userRole == "ADMIN" || userRole == "SUPER_ADMIN" {
		ws.Global.Subscribe(cl, "admin:emergencies")
	}

	// Auto-subscribe to rooms requested via query param
	if rooms := c.Query("rooms"); rooms != "" {
		for _, room := range strings.Split(rooms, ",") {
			room = strings.TrimSpace(room)
			if room != "" {
				ws.Global.Subscribe(cl, room)
			}
		}
	}

	ctx := c.Request.Context()

	// Heartbeat: send ping every 30 s
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				ws.Global.Broadcast("user:"+userID, ws.MsgPing, gin.H{"ts": time.Now().Unix()})
			case <-ctx.Done():
				return
			}
		}
	}()

	// Read loop — handle subscribe/unsubscribe frames from client
	for {
		var frame struct {
			Type string `json:"type"`
			Room string `json:"room"`
		}
		if err := wsjson.Read(ctx, conn, &frame); err != nil {
			if websocket.CloseStatus(err) == websocket.StatusNormalClosure ||
				websocket.CloseStatus(err) == websocket.StatusGoingAway ||
				ctx.Err() != nil {
				break
			}
			log.Printf("[WS] read error user=%s: %v", userID, err)
			break
		}
		switch frame.Type {
		case "subscribe":
			if frame.Room != "" {
				ws.Global.Subscribe(cl, frame.Room)
				log.Printf("[WS] user=%s subscribed to %s", userID, frame.Room)
			}
		case "ping":
			_ = wsjson.Write(context.Background(), conn, gin.H{"type": "pong", "ts": time.Now().Unix()})
		}
	}

	conn.Close(websocket.StatusNormalClosure, "bye")
}

// BroadcastEmergencyStatus is called by emergency handlers to push status changes.
func BroadcastEmergencyStatus(emergencyID string, payload interface{}) {
	ws.Global.Broadcast("emergency:"+emergencyID, ws.MsgEmergencyStatus, payload)
	ws.Global.Broadcast("admin:emergencies", ws.MsgEmergencyStatus, payload)
}

// BroadcastNewEmergency notifies admins of a freshly triggered SOS.
func BroadcastNewEmergency(payload interface{}) {
	ws.Global.Broadcast("admin:emergencies", ws.MsgEmergencyTriggered, payload)
}

// BroadcastAmbulanceLocation pushes a live ambulance position to the emergency room.
func BroadcastAmbulanceLocation(emergencyID string, payload interface{}) {
	ws.Global.Broadcast("emergency:"+emergencyID, ws.MsgAmbulanceLocation, payload)
}

