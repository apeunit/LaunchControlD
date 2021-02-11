// GENERATED BY THE COMMAND ABOVE; DO NOT EDIT
// This file was generated by swaggo/swag

package api

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/alecthomas/template"
	"github.com/swaggo/swag"
)

var doc = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{.Description}}",
        "title": "{{.Title}}",
        "contact": {
            "name": "API Support",
            "email": "u2467@apeunit.com"
        },
        "license": {
            "name": "MIT"
        },
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/status": {
            "get": {
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "health"
                ],
                "summary": "Healthcheck and version endpoint",
                "responses": {
                    "200": {
                        "description": "API Status",
                        "schema": {
                            "$ref": "#/definitions/server.APIStatus"
                        }
                    }
                }
            }
        },
        "/v1/auth/login": {
            "post": {
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "auth"
                ],
                "summary": "Login to the API",
                "parameters": [
                    {
                        "description": "Login credentials",
                        "name": "-",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/server.UserCredentials"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "API Reply",
                        "schema": {
                            "$ref": "#/definitions/server.APIReply"
                        }
                    }
                }
            }
        },
        "/v1/auth/logout": {
            "post": {
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "auth"
                ],
                "summary": "Logout from the system",
                "responses": {
                    "200": {
                        "description": "API Reply",
                        "schema": {
                            "$ref": "#/definitions/server.APIReply"
                        }
                    }
                }
            }
        },
        "/v1/auth/register": {
            "post": {
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "auth"
                ],
                "summary": "Register an API account",
                "parameters": [
                    {
                        "description": "Registration credentials",
                        "name": "-",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/server.UserCredentials"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "API Reply",
                        "schema": {
                            "$ref": "#/definitions/server.APIReply"
                        }
                    }
                }
            }
        },
        "/v1/events": {
            "get": {
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "event"
                ],
                "summary": "Retrieve a list of events",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/server.APIEvent"
                            }
                        }
                    }
                }
            },
            "post": {
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "event"
                ],
                "summary": "Create an event",
                "parameters": [
                    {
                        "description": "Event Request",
                        "name": "-",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/model.EventRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "API Reply",
                        "schema": {
                            "$ref": "#/definitions/server.APIReply"
                        }
                    }
                }
            }
        },
        "/v1/events/{id}": {
            "get": {
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "event"
                ],
                "summary": "Retrieve an event",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Event ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/server.APIEvent"
                        }
                    }
                }
            },
            "delete": {
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "event"
                ],
                "summary": "Destroy an event and associated resources",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Event ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/server.APIEvent"
                        }
                    }
                }
            }
        },
        "/v1/events/{id}/deploy": {
            "put": {
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "event"
                ],
                "summary": "Provision the insfrastructure and deploy the event",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Event ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/server.APIEvent"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "model.EventRequest": {
            "type": "object",
            "properties": {
                "genesis_accounts": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/model.GenesisAccount"
                    }
                },
                "owner": {
                    "type": "string"
                },
                "payload": {
                    "$ref": "#/definitions/model.PayloadLocation"
                },
                "provider": {
                    "type": "string"
                },
                "token_symbol": {
                    "type": "string"
                }
            }
        },
        "model.GenesisAccount": {
            "type": "object",
            "properties": {
                "faucet": {
                    "type": "boolean"
                },
                "genesis_balance": {
                    "type": "string"
                },
                "name": {
                    "type": "string"
                },
                "validator": {
                    "type": "boolean"
                }
            }
        },
        "model.PayloadLocation": {
            "type": "object",
            "properties": {
                "binary_path": {
                    "type": "string"
                },
                "binary_url": {
                    "type": "string"
                },
                "cli_path": {
                    "type": "string"
                },
                "daemon_path": {
                    "type": "string"
                },
                "docker_image": {
                    "type": "string"
                }
            }
        },
        "server.APIAccount": {
            "type": "object",
            "properties": {
                "address": {
                    "type": "string"
                },
                "faucet": {
                    "type": "boolean"
                },
                "genesis_balance": {
                    "type": "string"
                },
                "name": {
                    "type": "string"
                },
                "validator": {
                    "type": "boolean"
                }
            }
        },
        "server.APIEvent": {
            "type": "object",
            "properties": {
                "accounts": {
                    "type": "object",
                    "additionalProperties": {
                        "$ref": "#/definitions/server.APIAccount"
                    }
                },
                "created_on": {
                    "type": "string"
                },
                "ends_on": {
                    "type": "string"
                },
                "id": {
                    "type": "string"
                },
                "owner": {
                    "description": "email address of the owner",
                    "type": "string"
                },
                "provider": {
                    "description": "provider for provisioning",
                    "type": "string"
                },
                "starts_on": {
                    "type": "string"
                },
                "state": {
                    "type": "object",
                    "additionalProperties": {
                        "$ref": "#/definitions/server.APIMachineConfig"
                    }
                },
                "token_symbol": {
                    "description": "token symbool",
                    "type": "string"
                }
            }
        },
        "server.APIMachineConfig": {
            "type": "object",
            "properties": {
                "IPAddress": {
                    "type": "string"
                },
                "MachineName": {
                    "type": "string"
                },
                "tendermint_node_id": {
                    "type": "string"
                }
            }
        },
        "server.APIReply": {
            "type": "object",
            "properties": {
                "code": {
                    "type": "integer"
                },
                "message": {
                    "type": "string"
                }
            }
        },
        "server.APIStatus": {
            "type": "object",
            "properties": {
                "status": {
                    "type": "string"
                },
                "uptime": {
                    "type": "string"
                },
                "version": {
                    "type": "string"
                }
            }
        },
        "server.UserCredentials": {
            "type": "object",
            "properties": {
                "email": {
                    "type": "string"
                },
                "pass": {
                    "type": "string"
                }
            }
        }
    }
}`

type swaggerInfo struct {
	Version     string
	Host        string
	BasePath    string
	Schemes     []string
	Title       string
	Description string
}

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = swaggerInfo{
	Version:     "1.0",
	Host:        "api.launch-control.eventivize.co",
	BasePath:    "/api",
	Schemes:     []string{},
	Title:       "LaunchControlD REST API",
	Description: "This are the documentation for the LaunchControlD REST API",
}

type s struct{}

func (s *s) ReadDoc() string {
	sInfo := SwaggerInfo
	sInfo.Description = strings.ReplaceAll(sInfo.Description, "\n", "\\n")

	t, err := template.New("swagger_info").Funcs(template.FuncMap{
		"marshal": func(v interface{}) string {
			a, _ := json.Marshal(v)
			return string(a)
		},
	}).Parse(doc)
	if err != nil {
		return doc
	}

	var tpl bytes.Buffer
	if err := t.Execute(&tpl, sInfo); err != nil {
		return doc
	}

	return tpl.String()
}

func init() {
	swag.Register(swag.Name, &s{})
}
