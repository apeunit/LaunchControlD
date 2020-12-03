package server

import (
	log "github.com/sirupsen/logrus"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

// ServeHTTP starts the http service
func ServeHTTP() (err error) {
	log.Info("starting http")

	app := fiber.New()
	// enable cors
	app.Use(cors.New())
	// TODO: use logrus for logging
	app.Use(logger.New())
	// API group
	api := app.Group("/api")
	// V1
	v1 := api.Group("/v1")
	// define the api
	// CREATE
	v1.Post("/event", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Hello World",
		})
	})
	// DELETE
	v1.Delete("/event", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Hello World",
		})
	})
	// GET
	v1.Get("event", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Hello World",
		})
	})
	// LIST
	v1.Get("/events", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Hello World",
		})
	})

	err = app.Listen(":3000")
	return
}
