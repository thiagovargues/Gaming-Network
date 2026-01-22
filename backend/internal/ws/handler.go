package ws

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"backend/internal/config"
	"backend/internal/http/middleware"
	"backend/internal/repo"

	"github.com/gorilla/websocket"
)

type Client struct {
	UserID int64
	Conn   *websocket.Conn
	Send   chan []byte
}

type incomingMessage struct {
	Type    string `json:"type"`
	ToUser  int64  `json:"to_user_id,omitempty"`
	GroupID int64  `json:"group_id,omitempty"`
	Text    string `json:"text"`
}

type outgoingMessage struct {
	Type      string `json:"type"`
	FromUser  int64  `json:"from_user_id,omitempty"`
	ToUser    int64  `json:"to_user_id,omitempty"`
	GroupID   int64  `json:"group_id,omitempty"`
	Text      string `json:"text"`
	CreatedAt string `json:"created_at"`
	Message   string `json:"message,omitempty"`
}

var upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}

func NewHandler(cfg config.Config, db *sql.DB) http.HandlerFunc {
	hub := NewHub()

	return func(w http.ResponseWriter, r *http.Request) {
		_ = cfg
		current, ok := middleware.CurrentUser(r)
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		client := &Client{UserID: current.ID, Conn: conn, Send: make(chan []byte, 32)}
		hub.Register(client)

		go writePump(client)
		readPump(db, hub, client)
	}
}

func readPump(db *sql.DB, hub *Hub, client *Client) {
	defer func() {
		hub.Unregister(client)
		_ = client.Conn.Close()
	}()

	client.Conn.SetReadLimit(4096)
	_ = client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	client.Conn.SetPongHandler(func(string) error {
		return client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	})

	for {
		_, data, err := client.Conn.ReadMessage()
		if err != nil {
			return
		}
		var msg incomingMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			sendError(client, "invalid message")
			continue
		}
		msg.Text = strings.TrimSpace(msg.Text)
		if msg.Text == "" {
			sendError(client, "text required")
			continue
		}

		switch msg.Type {
		case "dm_send":
			handleDM(db, hub, client, msg)
		case "group_send":
			handleGroup(db, hub, client, msg)
		default:
			sendError(client, "unsupported type")
		}
	}
}

func writePump(client *Client) {
	defer func() {
		_ = client.Conn.Close()
	}()
	for msg := range client.Send {
		_ = client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
		if err := client.Conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			return
		}
	}
}

func handleDM(db *sql.DB, hub *Hub, client *Client, msg incomingMessage) {
	if msg.ToUser <= 0 {
		sendError(client, "to_user_id required")
		return
	}
	allowed, err := canDM(db, client.UserID, msg.ToUser)
	if err != nil || !allowed {
		sendError(client, "dm not allowed")
		return
	}
	created, err := repo.SaveDM(contextBackground(), db, client.UserID, msg.ToUser, msg.Text)
	if err != nil {
		sendError(client, "dm failed")
		return
	}

	payload := outgoingMessage{
		Type:      "dm_new",
		FromUser:  client.UserID,
		ToUser:    msg.ToUser,
		Text:      msg.Text,
		CreatedAt: created.CreatedAt,
	}
	data, _ := json.Marshal(payload)

	if shouldDeliverDM(db, client.UserID, msg.ToUser) {
		hub.SendToUser(msg.ToUser, data)
	}
	hub.SendToUser(client.UserID, data)
}

func handleGroup(db *sql.DB, hub *Hub, client *Client, msg incomingMessage) {
	if msg.GroupID <= 0 {
		sendError(client, "group_id required")
		return
	}
	member, err := repo.IsGroupMember(contextBackground(), db, msg.GroupID, client.UserID)
	if err != nil || !member {
		sendError(client, "not a member")
		return
	}
	created, err := repo.SaveGroupMessage(contextBackground(), db, msg.GroupID, client.UserID, msg.Text)
	if err != nil {
		sendError(client, "group message failed")
		return
	}
	payload := outgoingMessage{
		Type:      "group_new",
		FromUser:  client.UserID,
		GroupID:   msg.GroupID,
		Text:      msg.Text,
		CreatedAt: created.CreatedAt,
	}
	data, _ := json.Marshal(payload)

	members, err := repo.ListGroupMembers(contextBackground(), db, msg.GroupID)
	if err != nil {
		sendError(client, "group message failed")
		return
	}
	for _, member := range members {
		hub.SendToUser(member.ID, data)
	}
}

func sendError(client *Client, message string) {
	payload := outgoingMessage{Type: "error", Message: message}
	data, _ := json.Marshal(payload)
	select {
	case client.Send <- data:
	default:
	}
}

func canDM(db *sql.DB, fromID, toID int64) (bool, error) {
	f1, err := repo.IsFollowing(contextBackground(), db, fromID, toID)
	if err != nil {
		return false, err
	}
	f2, err := repo.IsFollowing(contextBackground(), db, toID, fromID)
	if err != nil {
		return false, err
	}
	return f1 || f2, nil
}

func shouldDeliverDM(db *sql.DB, fromID, toID int64) bool {
	follows, err := repo.IsFollowing(contextBackground(), db, toID, fromID)
	if err != nil {
		return false
	}
	if follows {
		return true
	}
	isPublic, _, err := repo.IsUserPublic(contextBackground(), db, toID)
	if err != nil {
		return false
	}
	return isPublic
}

func contextBackground() context.Context {
	return context.Background()
}
