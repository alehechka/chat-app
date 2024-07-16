package handlers

import (
	"chat-app/pkg/sockets"
	"fmt"
	"log/slog"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func HandleWebSockets(app *fiber.App) {
	app.Use("/ws", func(c *fiber.Ctx) error {
		// IsWebSocketUpgrade returns true if the client
		// requested upgrade to the WebSocket protocol.
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	app.Get("/ws/room/:id", websocket.New(func(c *websocket.Conn) {
		logger := slog.With()

		roomId, err := uuid.Parse(c.Params("id"))
		if err != nil {
			logger.Error("failed to parse roomId", slog.String("err", err.Error()))
			return
		}
		logger = logger.With(slog.String("roomId", roomId.String()))

		userId, err := uuid.Parse(c.Cookies("userId"))
		if err != nil {
			logger.Error("failed to parse userId", slog.String("err", err.Error()))
			return
		}
		logger = logger.With(slog.String("userId", userId.String()))

		conn := &sockets.Connection{Name: c.Cookies("username", "Anon"), Conn: c}
		if err := sockets.Pool.RegisterUser(roomId, userId, conn); err != nil {
			logger.Error("failed to register user", slog.String("err", err.Error()))
			return
		}

		// websocket.Conn bindings https://pkg.go.dev/github.com/fasthttp/websocket?tab=doc#pkg-index
		var (
			mt  int
			msg []byte
		)
		for {
			if mt, msg, err = c.ReadMessage(); err != nil {
				if err := sockets.Pool.BroadcastToRoom(roomId, userId, websocket.TextMessage, []byte(fmt.Sprintf("%s has left the room.", conn.Name))); err != nil {
					logger.Error("failed to broadcast exit message", slog.String("err", err.Error()))
				}
				break
			}

			if err = sockets.Pool.BroadcastToRoom(roomId, userId, mt, msg); err != nil {
				logger.Error("failed to broadcast message", slog.String("err", err.Error()), slog.String("msg", string(msg)))
				break
			}
		}
	}))
}
