package main

import (
	"chat-app/app/handlers"
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/urfave/cli/v2"
)

func main() {
	if err := app().Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

const (
	ArgPort = "port"
)

// App represents the CLI application
func app() *cli.App {
	app := cli.NewApp()
	app.Name = "chat-app"
	app.Usage = "Real-time Chat Application"
	app.Flags = []cli.Flag{
		&cli.IntFlag{
			Name:    ArgPort,
			Value:   3000,
			EnvVars: []string{"HTTP_LISTEN_ADDR", "PORT"},
		},
	}
	app.Action = chatApp

	return app
}

func chatApp(ctx *cli.Context) error {
	port := ctx.Int(ArgPort)

	router := fiber.New()
	handlers.HandleLanding(router)
	handlers.HandleWebSockets(router)

	return router.Listen(fmt.Sprintf(":%d", port))
}
