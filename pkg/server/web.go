package server

import (
	"github.com/apeunit/LaunchControlD/pkg/config"
	"github.com/apeunit/LaunchControlD/pkg/lctrld"
	"github.com/apeunit/LaunchControlD/pkg/model"
	log "github.com/sirupsen/logrus"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

// ServeHTTP starts the http service
func ServeHTTP(settings config.Schema) (err error) {
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

		e, err := model.ParseEventRequest(c.Body(), settings.DefaultPayloadLocation, settings.Web.DefaultProvider)
		if err != nil {
			c.JSON(fiber.Map{
				"error": err,
			})
		}

		log.Debug("%#v\n", e)
		event := model.NewEvent(e.TokenSymbol, e.Owner, "virtualbox", e.GenesisAccounts, e.PayloadLocation)
		err = lctrld.CreateEvent(settings, event)
		log.Debug("Creating event %#v\n", event)
		if err != nil {
			c.JSON(fiber.Map{
				"error": err,
			})
		}

		dmc := lctrld.NewDockerMachineConfig(settings, event.ID())
		event, err = lctrld.Provision(settings, event, lctrld.RunCommand, dmc)
		if err != nil {
			c.JSON(fiber.Map{
				"error": err,
			})
		}
		if err = lctrld.StoreEvent(settings, event); err != nil {
			c.JSON(fiber.Map{
				"error": err,
			})
		}
		// happy ending
		return c.JSON(fiber.Map{
			"id": event.ID(),
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
		events, err := lctrld.ListEvents(settings)
		if err != nil {
			c.JSON(fiber.ErrInternalServerError)
		}
		return c.JSON(events)
	})

	err = app.Listen(settings.Web.ListenAddress)
	return
}
