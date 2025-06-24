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

// RouteRequest define el cuerpo de la petición para generar una ruta.
type RouteRequest struct {
	StartPoint domain.Point    `json:"start_point" binding:"required"`
	Statuses   []domain.Status `json:"statuses" binding:"required"`
}

type UpsertContainerRequest struct {
	Latitude       float64 `json:"latitude" binding:"required,latitude"`
	Longitude      float64 `json:"longitude" binding:"required,longitude"`
	CapacityLiters int     `json:"capacity_liters" binding:"required,gt=0"`
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
	router.POST("/routes", h.CreateRoute)
	router.POST("/containers", h.CreateContainer)
	router.GET("/containers/:id", h.GetContainerByID)
	router.PUT("/containers/:id", h.UpdateContainer)
	router.DELETE("/containers/:id", h.DeleteContainer)
	router.GET("/containers/:id/readings", h.GetReadingsByContainerID)
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

// CreateRoute maneja la generación de una ruta de recogida optimizada.
// @Summary      Genera una ruta de recogida
// @Description  Calcula una ruta óptima para visitar contenedores basados en su estado.
// @Tags         Routes
// @Accept       json
// @Produce      json
// @Param        routeRequest body      RouteRequest      true  "Parámetros para la generación de la ruta"
// @Success      200          {object}  []domain.Container "La ruta optimizada como una lista ordenada de contenedores"
// @Failure      400          {object}  map[string]string "Petición inválida o datos incorrectos"
// @Failure      500          {object}  map[string]string "Error interno del servidor"
// @Router       /routes [post]
func (h *Handler) CreateRoute(c *gin.Context) {
	var req RouteRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cuerpo de la petición inválido: " + err.Error()})
		return
	}

	route, err := h.service.GenerateRoute(c.Request.Context(), req.StartPoint, req.Statuses)
	if err != nil {
		fmt.Printf("Error al generar la ruta: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudo generar la ruta"})
		return
	}

	c.JSON(http.StatusOK, route)
}

// @Summary      Crea un nuevo contenedor
// @Description  Registra un nuevo contenedor en el sistema con su ubicación y capacidad.
// @Tags         Containers
// @Accept       json
// @Produce      json
// @Param        container  body      UpsertContainerRequest  true  "Datos del contenedor a crear"
// @Success      201        {object}  domain.Container        "Contenedor creado exitosamente"
// @Failure      400        {object}  map[string]string       "Petición inválida o datos incorrectos"
// @Failure      500        {object}  map[string]string       "Error interno del servidor"
// @Router       /containers [post]
func (h *Handler) CreateContainer(c *gin.Context) {
	var req UpsertContainerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	newContainer := domain.Container{
		Location:       domain.Point{Latitude: req.Latitude, Longitude: req.Longitude},
		CapacityLiters: req.CapacityLiters,
	}

	created, err := h.service.CreateContainer(c.Request.Context(), newContainer)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudo crear el contenedor"})
		return
	}

	c.JSON(http.StatusCreated, created)
}

// @Summary      Obtiene un contenedor por su ID
// @Description  Devuelve la información detallada de un único contenedor.
// @Tags         Containers
// @Produce      json
// @Param        id   path      string  true  "ID del Contenedor (UUID)"
// @Success      200  {object}  domain.Container
// @Failure      404  {object}  map[string]string  "Contenedor no encontrado"
// @Failure      500  {object}  map[string]string  "Error interno del servidor"
// @Router       /containers/{id} [get]
func (h *Handler) GetContainerByID(c *gin.Context) {
	id := c.Param("id")
	container, err := h.service.GetContainerByID(c.Request.Context(), id)
	if err != nil {
		if err.Error() == "contenedor no encontrado" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al buscar el contenedor"})
		}
		return
	}
	c.JSON(http.StatusOK, container)
}

// @Summary      Actualiza un contenedor
// @Description  Actualiza la ubicación y/o la capacidad de un contenedor existente.
// @Tags         Containers
// @Accept       json
// @Produce      json
// @Param        id         path      string                  true  "ID del Contenedor (UUID)"
// @Param        container  body      UpsertContainerRequest  true  "Nuevos datos del contenedor"
// @Success      200        {object}  map[string]string       "Contenedor actualizado exitosamente"
// @Failure      400        {object}  map[string]string       "Petición inválida o datos incorrectos"
// @Failure      500        {object}  map[string]string       "Error interno del servidor"
// @Router       /containers/{id} [put]
func (h *Handler) UpdateContainer(c *gin.Context) {
	id := c.Param("id")
	var req UpsertContainerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	container := domain.Container{
		ID:             id,
		Location:       domain.Point{Latitude: req.Latitude, Longitude: req.Longitude},
		CapacityLiters: req.CapacityLiters,
	}

	if err := h.service.UpdateContainer(c.Request.Context(), container); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudo actualizar el contenedor"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Contenedor actualizado exitosamente"})
}

// @Summary      Elimina un contenedor
// @Description  Elimina un contenedor y todas sus lecturas asociadas del sistema.
// @Tags         Containers
// @Param        id   path      string  true  "ID del Contenedor (UUID)"
// @Success      204  "Sin contenido"
// @Failure      500  {object}  map[string]string "Error interno del servidor"
// @Router       /containers/{id} [delete]
func (h *Handler) DeleteContainer(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.DeleteContainer(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudo eliminar el contenedor"})
		return
	}
	c.Status(http.StatusNoContent)
}

// @Summary      Obtiene el historial de lecturas de un contenedor
// @Description  Devuelve una lista de las últimas N lecturas de sensor para un contenedor específico.
// @Tags         Containers
// @Produce      json
// @Param        id      path      string  true   "ID del Contenedor (UUID)"
// @Param        limit   query     int     false  "Número máximo de lecturas a devolver (por defecto 50)"
// @Success      200     {object}  []domain.Reading
// @Failure      500     {object}  map[string]string "Error interno del servidor"
// @Router       /containers/{id}/readings [get]
func (h *Handler) GetReadingsByContainerID(c *gin.Context) {
	id := c.Param("id")
	// En un caso real, el límite podría venir como un query param:
	// limitStr := c.DefaultQuery("limit", "50")
	// limit, _ := strconv.Atoi(limitStr)
	readings, err := h.service.GetReadingsForContainer(c.Request.Context(), id, 50)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudieron obtener las lecturas"})
		return
	}
	c.JSON(http.StatusOK, readings)
}
