package server

import (
	"github.com/apeunit/LaunchControlD/pkg/config"
	"github.com/apeunit/LaunchControlD/pkg/lctrld"
	"github.com/apeunit/LaunchControlD/pkg/model"
	log "github.com/sirupsen/logrus"

	swagger "github.com/arsmn/fiber-swagger/v2"
	// swagger generated documentation
	_ "github.com/apeunit/LaunchControlD/api"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

var (
	appSettings config.Schema
)

// ServeHTTP starts the http service
// @title LaunchControlD REST API
// @version 1.0
// @description This are the documentation for the LaunchControlD REST API
// @contact.name API Support
// @contact.email u2467@apeunit.com
// @license.name MIT
// @host localhost:2012
// @BasePath /api/v1
func ServeHTTP(settings config.Schema) (err error) {
	log.Info("starting http")
	// make settings available to the other functions
	appSettings = settings
	// setup the web framework
	app := fiber.New()
	// enable cors
	app.Use(cors.New())
	// TODO: use logrus for logging
	app.Use(logger.New())
	// handle swagger routes
	app.Get("/swagger/*", swagger.Handler) // default
	// API group
	api := app.Group("/api")
	v1 := api.Group("/v1")
	// define the api
	v1.Post("/events", eventCreate)
	v1.Put("/events/:eventID/deploy", eventDeploy)
	v1.Delete("/events/:eventID", deleteEvent)
	v1.Get("/events/:eventID", getEvent)
	v1.Get("/events", listEvents)
	// run the web server
	err = app.Listen(settings.Web.ListenAddress)
	return
}

// eventCreate godoc
// @Summary Create an event
// @Tags event
// @Accept  json
// @Produce  json
// @Param - body model.EventRequest true "Event Request"
// @Success 200 {object} model.Event
// @Router /events [post]
func eventCreate(c *fiber.Ctx) error {
	er, err := model.ParseEventRequest(c.Body(),
		appSettings.DefaultPayloadLocation,
		//TODO: provide a better way to set defaults
		appSettings.Web.DefaultProvider,
	)
	if err != nil {
		c.JSON(fiber.Map{
			"error": err,
		})
	}
	log.Debug("event request: %#v\n", er)
	// now create a new event
	event := model.NewEvent(er.TokenSymbol,
		er.Owner,
		er.Provider,
		er.GenesisAccounts,
		er.PayloadLocation,
	)
	err = lctrld.CreateEvent(appSettings, event)
	log.Debug("Creating event %#v\n", event)
	if err != nil {
		c.JSON(fiber.Map{
			"error": err,
		})
	}
	// store event
	if err = lctrld.StoreEvent(appSettings, event); err != nil {
		c.JSON(fiber.Map{
			"error": err,
		})
	}
	// happy ending
	return c.JSON(fiber.Map{
		"id": event.ID(),
	})
}

// @Summary Provision the insfrastructure and deploy the event
// @Tags event
// @Accept  json
// @Produce  json
// @Param id path string true "Event ID"
// @Success 200 {object} model.Event
// @Router /events/{id}/deploy [put]
func eventDeploy(c *fiber.Ctx) error {
	eventID := c.Params("eventID")
	event, err := lctrld.GetEventByID(appSettings, eventID)
	if err != nil {
		return c.JSON(fiber.ErrNotFound)
	}

	dmc := lctrld.NewDockerMachineConfig(appSettings, event.ID())
	_, err = lctrld.Provision(appSettings, &event, lctrld.RunCommand, dmc)
	if err != nil {
		return c.JSON(fiber.Map{
			"error": err,
		})
	}
	return c.JSON(fiber.ErrTeapot)
}

// @Summary Destroy an event and associated resources
// @Tags event
// @Accept  json
// @Produce  json
// @Param id path string true "Event ID"
// @Success 200 {object} model.Event
// @Router /events/{id} [delete]
func deleteEvent(c *fiber.Ctx) error {
	eventID := c.Params("eventID")
	evt, err := lctrld.GetEventByID(appSettings, eventID)
	if err != nil {
		return c.JSON(fiber.ErrNotFound)
	}
	err = lctrld.DestroyEvent(appSettings, &evt, lctrld.RunCommand)
	if err != nil {
		return c.JSON(fiber.ErrInternalServerError)
	}
	return c.JSON(evt)
}

// @Summary Retrieve an event
// @Tags event
// @Accept  json
// @Produce  json
// @Param id path string true "Event ID"
// @Success 200 {object} model.Event
// @Router /events/{id} [get]
func getEvent(c *fiber.Ctx) error {
	eventID := c.Params("eventID")
	evt, err := lctrld.GetEventByID(appSettings, eventID)
	if err != nil {
		return c.JSON(fiber.ErrNotFound)
	}
	return c.JSON(evt)
}

// @Summary Retrieve a list of events
// @Tags event
// @Accept  json
// @Produce  json
// @Success 200 {array} model.Event
// @Router /events [get]
func listEvents(c *fiber.Ctx) error {
	events, err := lctrld.ListEvents(appSettings)
	if err != nil {
		return c.JSON(fiber.ErrInternalServerError)
	}
	return c.JSON(events)
}
