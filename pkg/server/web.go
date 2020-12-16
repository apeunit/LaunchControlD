package server

import (
	"github.com/apeunit/LaunchControlD/pkg/config"
	"github.com/apeunit/LaunchControlD/pkg/lctrld"
	"github.com/apeunit/LaunchControlD/pkg/model"
	"github.com/apeunit/LaunchControlD/pkg/utils"
	log "github.com/sirupsen/logrus"

	swagger "github.com/arsmn/fiber-swagger/v2"
	// swagger generated documentation
	_ "github.com/apeunit/LaunchControlD/api"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/session"
)

const (
	sessionKeyUserHash = "user_hash"
	headerAuthToken    = "X-LCTRLD-TOKEN"
)

var (
	appSettings config.Schema
	usersDb     *UsersDB
)

// UserCredentials the input user credential for authentication
type UserCredentials struct {
	Email string `json:"email,omitempty"`
	Pass  string `json:"pass,omitempty"`
}

// APIReply a reply from the API
type APIReply struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

// APIReplyOK returns an 200 reply
func APIReplyOK(m string) APIReply {
	return APIReply{
		Status:  fiber.StatusOK,
		Message: m,
	}
}

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
	usersDb, err = NewUserDB(utils.GetPath(settings.Workspace, settings.Web.UsersDbFile))
	if err != nil {
		return
	}
	// setup the web framework
	app := fiber.New()
	// enable cors
	app.Use(cors.New())
	// TODO: use logrus for logging
	app.Use(logger.New())
	// session management

	// handle swagger routes
	app.Get("/swagger/*", swagger.Handler) // default
	// API group
	api := app.Group("/api")
	v1 := api.Group("/v1")
	// define the api
	v1.Post("/auth/login", login)
	v1.Post("/auth/logout", logout)
	v1.Post("/auth/register", register)
	// events api
	events := v1.Group("/events")
	// add the authorization middleware
	events.Use(auth)
	// register the routes
	events.Post("/", eventCreate)
	events.Put("/:eventID/deploy", eventDeploy)
	events.Delete("/:eventID", deleteEvent)
	events.Get("/:eventID", getEvent)
	events.Get("/", listEvents)
	// run the web server
	err = app.Listen(settings.Web.ListenAddress)
	return
}

// auth is a middleware specific for the events
func auth(c *fiber.Ctx) error {
	//	TODO: sadly we cannot easily propagate the email
	_, err := usersDb.IsTokenAuthorized(c.Get(headerAuthToken))
	if err != nil {
		return c.JSON(fiber.ErrUnauthorized)
	}
	// Go to next middleware:
	return c.Next()
}

func isLoggedIn(s *session.Session) bool {
	if s.Get(sessionKeyUserHash) == nil {
		return false
	}
	return true
}

// @Summary Login to the API
// @Tags auth, session
// @Accept  json
// @Produce  json
// @Param - body UserCredentials true "Login credentials"
// @Success 200 {object} APIReply "API Reply"
// @Router /auth/login [post]
func login(c *fiber.Ctx) error {
	// retrieve the credentials
	var credentials UserCredentials
	err := c.BodyParser(&credentials)
	if err != nil {
		log.Error(err)
		return c.JSON(fiber.ErrBadRequest)
	}
	// validate the credentials
	token, err := usersDb.IsAuthorized(credentials.Email, credentials.Pass)
	if err != nil {
		return c.JSON(fiber.ErrUnauthorized)
	}
	// reply token in headers
	c.Set(headerAuthToken, token)
	return c.JSON(APIReplyOK(token))
}

// @Summary Logout from the system
// @Tags auth, session
// @Accept  json
// @Produce  json
// @Success 200 {string} string "ok"
// @Router /auth/logout [post]
func logout(c *fiber.Ctx) error {
	// get session from storage
	usersDb.DropToken(c.Get(headerAuthToken))
	return c.JSON(APIReplyOK("ok"))
}

// @Summary Register an API account
// @Tags auth, session
// @Accept  json
// @Produce  json
// @Param - body UserCredentials true "Registration credentials"
// @Success 200 {string} string "ok"
// @Router /auth/register [post]
func register(c *fiber.Ctx) error {
	// retrieve the credentials
	var credentials UserCredentials
	err := c.BodyParser(&credentials)
	if err != nil {
		log.Error(err)
		return c.JSON(fiber.ErrBadRequest)
	}
	log.Info("credentials are", credentials)
	// register the new user
	err = usersDb.RegisterUser(credentials.Email, credentials.Pass)
	if err != nil {
		return c.JSON(fiber.ErrExpectationFailed)
	}
	log.Info("ready to go")
	return c.JSON(APIReplyOK("ok"))
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
	return c.JSON(APIReplyOK(event.ID()))
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
