package container

import (
	"fmt"
	"net/http"
	"smart-waste-management/internal/domain"

	"github.com/gin-gonic/gin"
)

// Handler maneja las peticiones HTTP para los recursos de contenedores.
// Depende de la interfaz del Servicio, no de su implementación.
type Handler struct {
	service Service
}

// NewHandler crea una nueva instancia del handler.
func NewHandler(s Service) *Handler {
	return &Handler{
		service: s,
	}
}

// RegisterRoutes registra todas las rutas de este handler en el router de Gin.
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	router.POST("/readings", h.CreateReading)
	router.GET("/containers", h.GetContainers)
}

// CreateReading maneja la creación de una nueva lectura de sensor.
// @Summary      Crea una nueva lectura de sensor
// @Description  Registra el nivel de llenado de un contenedor en un momento dado.
// @Tags         Ingest
// @Accept       json
// @Produce      json
// @Param        reading  body      domain.Reading  true  "Datos de la lectura"
// @Success      202  {object}  map[string]string "Lectura aceptada para procesamiento"
// @Failure      400  {object}  map[string]string "Petición inválida o datos incorrectos"
// @Failure      500  {object}  map[string]string "Error interno del servidor"
// @Router       /readings [post]
func (h *Handler) CreateReading(c *gin.Context) {
	var reading domain.Reading

	// 1. Decodificar y validar estructuralmente el JSON del cuerpo de la petición.
	if err := c.ShouldBindJSON(&reading); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cuerpo de la petición inválido: " + err.Error()})
		return
	}

	// 2. Llamar a la capa de servicio para procesar la lógica de negocio.
	err := h.service.ProcessNewReading(c.Request.Context(), reading)
	if err != nil {
		// 4. Manejar errores del servicio.
		// Aquí podríamos tener una lógica más sofisticada para mapear tipos de error a códigos de estado.
		fmt.Printf("Error al procesar la lectura: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudo procesar la lectura"})
		return
	}

	// 5. Enviar la respuesta. 202 Accepted es semánticamente correcto para una ingesta de datos asíncrona.
	c.JSON(http.StatusAccepted, gin.H{"message": "Lectura aceptada"})
}

// GetContainers maneja la obtención de todos los contenedores.
// @Summary      Obtiene todos los contenedores
// @Description  Devuelve una lista de todos los contenedores registrados con su estado actual.
// @Tags         Containers
// @Produce      json
// @Success      200  {object}  []domain.Container
// @Failure      500  {object}  map[string]string "Error interno del servidor"
// @Router       /containers [get]
func (h *Handler) GetContainers(c *gin.Context) {
	// 2. Llamar al servicio.
	containers, err := h.service.GetAllContainers(c.Request.Context())
	if err != nil {
		fmt.Printf("Error al obtener los contenedores: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudo obtener la lista de contenedores"})
		return
	}

	// 5. Enviar la respuesta.
	c.JSON(http.StatusOK, containers)
}
