package handler

import "github.com/gofiber/fiber/v2"

func InitRoutes(app *fiber.App) {
	app.Get("/helloworld", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

}
