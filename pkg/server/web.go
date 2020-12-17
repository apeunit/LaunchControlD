package server

import (
	"fmt"
	"net/http"
	"strings"
	"time"

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
	Status  int    `json:"code"`
	Message string `json:"message"`
}

// APIStatus hold the status of the API
type APIStatus struct {
	Status  string `json:"status,omitempty"`
	Version string `json:"version,omitempty"`
	Uptime  string `json:"uptime,omitempty"`
}

// APIReplyOK returns an 200 reply
func APIReplyOK(m string) APIReply {
	return APIReply{
		Status:  http.StatusOK,
		Message: m,
	}
}

// APIReplyErr error reply
func APIReplyErr(code int, m string) APIReply {
	return APIReply{
		Status:  code,
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
// @host api.launch-control.eventivize.co
// @BasePath /api
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

	// root url
	app.Get("/", func(c *fiber.Ctx) error { return c.JSON(fiber.ErrTeapot) })
	// handle swagger routes
	app.Get("/swagger/*", swagger.Handler) // default
	// API group
	api := app.Group("/api")
	api.Get("/status", status)
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

// retrieve the email of the authenticated user
func getAuthEmail(c *fiber.Ctx) (email string, err error) {
	email, err = usersDb.GetEmailFromToken(c.Get(headerAuthToken))
	return
}

// @Summary Healthcheck and version endpoint
// @Tags health
// @Produce  json
// @Success 200 {object} APIStatus "API Status"
// @Router /status [get]
func status(c *fiber.Ctx) error {
	return c.JSON(APIStatus{
		Status:  "OK",
		Version: appSettings.RuntimeVersion,
		Uptime:  fmt.Sprint(time.Since(appSettings.RuntimeStartedAt)),
	})
}

// @Summary Login to the API
// @Tags auth
// @Accept  json
// @Produce  json
// @Param - body UserCredentials true "Login credentials"
// @Success 200 {object} APIReply "API Reply"
// @Router /v1/auth/login [post]
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
// @Tags auth
// @Accept  json
// @Produce  json
// @Success 200 {string} string "ok"
// @Router /v1/auth/logout [post]
func logout(c *fiber.Ctx) error {
	// get session from storage
	usersDb.DropToken(c.Get(headerAuthToken))
	return c.JSON(APIReplyOK("ok"))
}

// @Summary Register an API account
// @Tags auth
// @Accept  json
// @Produce  json
// @Param - body UserCredentials true "Registration credentials"
// @Success 200 {string} string "ok"
// @Router /v1/auth/register [post]
func register(c *fiber.Ctx) error {
	// retrieve the credentials
	var credentials UserCredentials
	err := c.BodyParser(&credentials)
	if err != nil {
		log.Error(err)
		return c.JSON(fiber.ErrBadRequest)
	}
	// register the new user
	err = usersDb.RegisterUser(credentials.Email, credentials.Pass)
	if err != nil {
		return c.JSON(fiber.ErrExpectationFailed)
	}
	return c.JSON(APIReplyOK("ok"))
}

// eventCreate godoc
// @Summary Create an event
// @Tags event
// @Accept  json
// @Produce  json
// @Param - body model.EventRequest true "Event Request"
// @Success 200 {object} model.Event
// @Router /v1/events [post]
func eventCreate(c *fiber.Ctx) error {
	// retrieve the owner email
	ownerEmail, err := getAuthEmail(c)
	//parse the event requests
	var er model.EventRequest
	c.BodyParser(&er)
	log.Debugf("REST: event request %#v", er)
	// TODO: find a better way for defaults
	er.Provider = appSettings.Web.DefaultProvider
	er.PayloadLocation = appSettings.DefaultPayloadLocation
	// override the owner
	er.Owner = ownerEmail
	// validate the event request
	if strings.TrimSpace(er.TokenSymbol) == "" {
		return c.JSON(fiber.ErrBadRequest)
	}
	// count the number of accounts
	if len(er.GenesisAccounts) == 0 {
		return c.JSON(fiber.ErrBadRequest)
	}
	// check that the names are set
	for _, g := range er.GenesisAccounts {
		if strings.TrimSpace(g.Name) == "" {
			return c.JSON(fiber.ErrBadRequest)
		}
	}
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
		return c.JSON(APIReplyErr(http.StatusInternalServerError, err.Error()))
	}
	// store event
	if err = lctrld.StoreEvent(appSettings, event); err != nil {
		return c.JSON(APIReplyErr(http.StatusInternalServerError, err.Error()))
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
// @Router /v1/events/{id}/deploy [put]
func eventDeploy(c *fiber.Ctx) error {
	eventID := c.Params("eventID")
	event, err := lctrld.GetEventByID(appSettings, eventID)
	if err != nil {
		return c.JSON(fiber.ErrNotFound)
	}

	dmc := lctrld.NewDockerMachineConfig(appSettings, event.ID())
	_, err = lctrld.Provision(appSettings, &event, lctrld.RunCommand, dmc)
	if err != nil {
		return c.JSON(APIReplyErr(http.StatusInternalServerError, err.Error()))
	}
	return c.JSON(fiber.ErrTeapot)
}

// @Summary Destroy an event and associated resources
// @Tags event
// @Accept  json
// @Produce  json
// @Param id path string true "Event ID"
// @Success 200 {object} model.Event
// @Router /v1/events/{id} [delete]
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
// @Router /v1/events/{id} [get]
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
// @Router /v1/events [get]
func listEvents(c *fiber.Ctx) error {
	events, err := lctrld.ListEvents(appSettings)
	if err != nil {
		return c.JSON(fiber.ErrInternalServerError)
	}
	return c.JSON(events)
}
