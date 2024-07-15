package main

import (
	"chat-app/views"
	"fmt"
	"log"
	"sync"

	"github.com/a-h/templ"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
)

type SocketRegistry struct {
	Mutex       sync.Mutex
	Connections map[string]*websocket.Conn
}

func NewSocketRegistry() *SocketRegistry {
	return &SocketRegistry{
		Connections: make(map[string]*websocket.Conn),
	}
}

func (r *SocketRegistry) Register(userId string, ws *websocket.Conn) {
	r.Mutex.Lock()
	defer r.Mutex.Unlock()

	r.Connections[userId] = ws
}

func (r *SocketRegistry) Unregister(userId string) {
	r.Mutex.Lock()
	defer r.Mutex.Unlock()

	if _, ok := r.Connections[userId]; ok {
		delete(r.Connections, userId)
	}
}

func (r *SocketRegistry) SendToOthers(userId string, mt int, msg []byte) error {
	for user, ws := range r.Connections {
		if user != userId {
			if err := ws.WriteMessage(mt, []byte(fmt.Sprintf("%s: %s", userId, msg))); err != nil {
				return err
			}
		}
	}
	return nil
}

var RegistryPool = NewSocketRegistry()

func main() {
	app := fiber.New()

	app.Use("/ws", func(c *fiber.Ctx) error {
		// IsWebSocketUpgrade returns true if the client
		// requested upgrade to the WebSocket protocol.
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	app.Get("/echo/:id", websocket.New(func(c *websocket.Conn) {
		// c.Locals is added to the *websocket.Conn
		id := c.Params("id")
		log.Println(c.Locals("allowed"))  // true
		log.Println(id)                   // 123
		log.Println(c.Query("v"))         // 1.0
		log.Println(c.Cookies("session")) // ""
		RegistryPool.Register(id, c)

		// websocket.Conn bindings https://pkg.go.dev/github.com/fasthttp/websocket?tab=doc#pkg-index
		var (
			mt  int
			msg []byte
			err error
		)
		for {
			if mt, msg, err = c.ReadMessage(); err != nil {
				log.Println("read:", err)
				break
			}
			log.Printf("recv: %s", msg)

			if err = RegistryPool.SendToOthers(id, mt, msg); err != nil {
				log.Println("write:", err)
				break
			}
		}

	}))

	app.Get("/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		return Render(c, views.EchoPage("ws://"+c.Hostname()+"/echo/"+id))
	})

	log.Fatal(app.Listen(":3000"))
	// Access the websocket server: ws://localhost:3000/ws/123?v=1.0
	// https://www.websocket.org/echo.html
}

func Render(c *fiber.Ctx, component templ.Component, options ...func(*templ.ComponentHandler)) error {
	componentHandler := templ.Handler(component)
	for _, o := range options {
		o(componentHandler)
	}
	return adaptor.HTTPHandler(componentHandler)(c)
}
