package handlers

import (
	"chat-app/app/views"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func HandleLanding(app *fiber.App) {
	app.Get("/", func(c *fiber.Ctx) error {
		if _, err := uuid.Parse(c.Cookies("userId")); err != nil {
			return c.Redirect("/login", http.StatusTemporaryRedirect)
		}

		return Render(c, views.LandingPage())
	})
}
