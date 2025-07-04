// Package docs Code generated by swaggo/swag. DO NOT EDIT
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "termsOfService": "http://swagger.io/terms/",
        "contact": {
            "name": "API Support",
            "url": "http://www.example.com/support",
            "email": "support@example.com"
        },
        "license": {
            "name": "Apache 2.0",
            "url": "http://www.apache.org/licenses/LICENSE-2.0.html"
        },
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/containers": {
            "get": {
                "description": "Devuelve una lista de todos los contenedores registrados con su estado actual.",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Containers"
                ],
                "summary": "Obtiene todos los contenedores",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/domain.Container"
                            }
                        }
                    },
                    "500": {
                        "description": "Error interno del servidor",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    }
                }
            },
            "post": {
                "description": "Registra un nuevo contenedor en el sistema con su ubicación y capacidad.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Containers"
                ],
                "summary": "Crea un nuevo contenedor",
                "parameters": [
                    {
                        "description": "Datos del contenedor a crear",
                        "name": "container",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/container.UpsertContainerRequest"
                        }
                    }
                ],
                "responses": {
                    "201": {
                        "description": "Contenedor creado exitosamente",
                        "schema": {
                            "$ref": "#/definitions/domain.Container"
                        }
                    },
                    "400": {
                        "description": "Petición inválida o datos incorrectos",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    },
                    "500": {
                        "description": "Error interno del servidor",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    }
                }
            }
        },
        "/containers/{id}": {
            "get": {
                "description": "Devuelve la información detallada de un único contenedor.",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Containers"
                ],
                "summary": "Obtiene un contenedor por su ID",
                "parameters": [
                    {
                        "type": "string",
                        "description": "ID del Contenedor (UUID)",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/domain.Container"
                        }
                    },
                    "404": {
                        "description": "Contenedor no encontrado",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    },
                    "500": {
                        "description": "Error interno del servidor",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    }
                }
            },
            "put": {
                "description": "Actualiza la ubicación y/o la capacidad de un contenedor existente.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Containers"
                ],
                "summary": "Actualiza un contenedor",
                "parameters": [
                    {
                        "type": "string",
                        "description": "ID del Contenedor (UUID)",
                        "name": "id",
                        "in": "path",
                        "required": true
                    },
                    {
                        "description": "Nuevos datos del contenedor",
                        "name": "container",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/container.UpsertContainerRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Contenedor actualizado exitosamente",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    },
                    "400": {
                        "description": "Petición inválida o datos incorrectos",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    },
                    "500": {
                        "description": "Error interno del servidor",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    }
                }
            },
            "delete": {
                "description": "Elimina un contenedor y todas sus lecturas asociadas del sistema.",
                "tags": [
                    "Containers"
                ],
                "summary": "Elimina un contenedor",
                "parameters": [
                    {
                        "type": "string",
                        "description": "ID del Contenedor (UUID)",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "204": {
                        "description": "Sin contenido"
                    },
                    "500": {
                        "description": "Error interno del servidor",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    }
                }
            }
        },
        "/containers/{id}/readings": {
            "get": {
                "description": "Devuelve una lista de las últimas N lecturas de sensor para un contenedor específico.",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Containers"
                ],
                "summary": "Obtiene el historial de lecturas de un contenedor",
                "parameters": [
                    {
                        "type": "string",
                        "description": "ID del Contenedor (UUID)",
                        "name": "id",
                        "in": "path",
                        "required": true
                    },
                    {
                        "type": "integer",
                        "description": "Número máximo de lecturas a devolver (por defecto 50)",
                        "name": "limit",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/domain.Reading"
                            }
                        }
                    },
                    "500": {
                        "description": "Error interno del servidor",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    }
                }
            }
        },
        "/readings": {
            "post": {
                "description": "Registra el nivel de llenado de un contenedor en un momento dado.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Ingest"
                ],
                "summary": "Crea una nueva lectura de sensor",
                "parameters": [
                    {
                        "description": "Datos de la lectura",
                        "name": "reading",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/domain.Reading"
                        }
                    }
                ],
                "responses": {
                    "202": {
                        "description": "Lectura aceptada para procesamiento",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    },
                    "400": {
                        "description": "Petición inválida o datos incorrectos",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    },
                    "500": {
                        "description": "Error interno del servidor",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    }
                }
            }
        },
        "/routes": {
            "post": {
                "description": "Calcula una ruta óptima para visitar contenedores basados en su estado.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Routes"
                ],
                "summary": "Genera una ruta de recogida",
                "parameters": [
                    {
                        "description": "Parámetros para la generación de la ruta",
                        "name": "routeRequest",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/container.RouteRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "La ruta optimizada como una lista ordenada de contenedores",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/domain.Container"
                            }
                        }
                    },
                    "400": {
                        "description": "Petición inválida o datos incorrectos",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    },
                    "500": {
                        "description": "Error interno del servidor",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "container.RouteRequest": {
            "type": "object",
            "required": [
                "start_point",
                "statuses"
            ],
            "properties": {
                "start_point": {
                    "$ref": "#/definitions/domain.Point"
                },
                "statuses": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/domain.Status"
                    }
                }
            }
        },
        "container.UpsertContainerRequest": {
            "type": "object",
            "required": [
                "capacity_liters",
                "latitude",
                "longitude"
            ],
            "properties": {
                "capacity_liters": {
                    "type": "integer"
                },
                "latitude": {
                    "type": "number"
                },
                "longitude": {
                    "type": "number"
                }
            }
        },
        "domain.Container": {
            "type": "object",
            "properties": {
                "capacity_liters": {
                    "type": "integer"
                },
                "created_at": {
                    "description": "--- CAMPOS ACTUALIZADOS ---\nEstos campos son gestionados por la base de datos y son cruciales para el tracking.",
                    "type": "string"
                },
                "id": {
                    "type": "string"
                },
                "last_fill_level": {
                    "type": "integer"
                },
                "last_updated": {
                    "type": "string"
                },
                "location": {
                    "$ref": "#/definitions/domain.Point"
                },
                "status": {
                    "description": "omitempty porque no se establece al crear",
                    "allOf": [
                        {
                            "$ref": "#/definitions/domain.Status"
                        }
                    ]
                },
                "updated_at": {
                    "type": "string"
                }
            }
        },
        "domain.Point": {
            "type": "object",
            "properties": {
                "latitude": {
                    "type": "number"
                },
                "longitude": {
                    "type": "number"
                }
            }
        },
        "domain.Reading": {
            "type": "object",
            "properties": {
                "container_id": {
                    "type": "string"
                },
                "fill_level": {
                    "type": "integer"
                },
                "timestamp": {
                    "type": "string"
                }
            }
        },
        "domain.Status": {
            "type": "string",
            "enum": [
                "low",
                "medium",
                "high"
            ],
            "x-enum-varnames": [
                "StatusLow",
                "StatusMedium",
                "StatusHigh"
            ]
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "1.0",
	Host:             "localhost:8080",
	BasePath:         "/api/v1",
	Schemes:          []string{},
	Title:            "Smart Waste Management API",
	Description:      "API para la gestión y monitorización en tiempo real de contenedores de basura.",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
