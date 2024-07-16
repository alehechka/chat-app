package sockets

import (
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"github.com/gofiber/contrib/websocket"
	"github.com/google/uuid"
)

type Registry struct {
	Mutex       sync.Mutex
	Connections map[uuid.UUID]map[uuid.UUID]*Connection
}

type Connection struct {
	Name string
	Conn *websocket.Conn
}

func NewRegistry() *Registry {
	return &Registry{
		Connections: make(map[uuid.UUID]map[uuid.UUID]*Connection),
	}
}

func (r *Registry) OpenRoom() uuid.UUID {
	r.Mutex.Lock()
	defer r.Mutex.Unlock()

	id := uuid.New()
	r.Connections[id] = make(map[uuid.UUID]*Connection)
	return id
}

func (r *Registry) CloseRoom(roomId uuid.UUID) error {
	r.Mutex.Lock()
	defer r.Mutex.Unlock()

	room, ok := r.Connections[roomId]
	if !ok || room == nil {
		return errors.New("room does not exist")
	}

	for userId, user := range room {
		if user != nil {
			user.Conn.Close()
		}
		delete(room, userId)
	}

	return nil
}

func (r *Registry) RegisterUser(roomId uuid.UUID, userId uuid.UUID, conn *Connection) error {
	r.Mutex.Lock()
	defer r.Mutex.Unlock()

	room, ok := r.Connections[roomId]
	if !ok || room == nil {
		return errors.New("room does not exist")
	}

	if user := room[userId]; user != nil && user.Conn != nil {
		user.Conn.Close()
	}
	room[userId] = conn

	return r.BroadcastToRoom(roomId, userId, websocket.TextMessage, []byte(fmt.Sprintf("%s has joined the room.", conn.Name)))
}

func (r *Registry) UnregisterUser(roomId uuid.UUID, userId uuid.UUID) error {
	r.Mutex.Lock()
	defer r.Mutex.Unlock()

	room, ok := r.Connections[roomId]
	if !ok || room == nil {
		return errors.New("room does not exist")
	}

	if user := room[userId]; user != nil && user.Conn != nil {
		user.Conn.Close()
	}
	delete(room, userId)

	if len(room) == 0 {
		return r.CloseRoom(roomId)
	}

	return nil
}

func (r *Registry) BroadcastToRoom(roomId uuid.UUID, userId uuid.UUID, mt int, msg []byte) error {
	fullMessage := []byte(fmt.Sprintf("%s: %s", userId, msg)) // render this with templ

	room, ok := r.Connections[roomId]
	if !ok || room == nil {
		return errors.New("room does not exist")
	}

	for user, conn := range room {
		if user != userId {
			if err := conn.Conn.WriteMessage(mt, fullMessage); err != nil {
				slog.Error("failed to broadcast message",
					slog.String("err", err.Error()),
					slog.String("msg", string(fullMessage)),
					slog.String("room", roomId.String()),
					slog.String("sender", userId.String()),
					slog.String("receiver", user.String()),
				)
			}
		}
	}

	return nil
}

var Pool = NewRegistry()
