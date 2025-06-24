package main

import (
	"log"
	"net/http"
	"os"
	"smart-waste-management/internal/container"
	"smart-waste-management/internal/platform/database"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	// --- Swagger ---
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	// Importante: el alias _ importa el paquete docs para que swag lo encuentre.
	// Asegúrate de que la ruta sea correcta según tu go.mod.
	_ "smart-waste-management/docs"
)

// @title        Smart Waste Management API
// @version      1.0
// @description  API para la gestión y monitorización en tiempo real de contenedores de basura.
// @termsOfService http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.example.com/support
// @contact.email  support@example.com

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host       localhost:8080
// @BasePath   /api/v1
func main() {
	// 1. Cargar configuración desde el fichero .env
	if err := godotenv.Load(); err != nil {
		log.Println("Advertencia: No se pudo cargar el fichero .env. Se usarán las variables de entorno del sistema.")
	}

	// 2. Establecer conexión con la base de datos
	db, err := database.NewDBConnection()
	if err != nil {
		log.Fatalf("FATAL: No se pudo conectar a la base de datos: %v", err)
	}
	defer db.Close() // Asegura que las conexiones se cierren al final de main

	// 3. "Cablear" las dependencias (Dependency Injection)
	// La cadena es: DB -> Repositorio -> Servicio -> Handler
	containerRepository := container.NewPostgresRepository(db)
	containerService := container.NewService(containerRepository)
	containerHandler := container.NewHandler(containerService)

	// 4. Configurar el router de Gin
	router := setupRouter(containerHandler)

	// 5. Arrancar el servidor HTTP
	apiPort := os.Getenv("API_PORT")
	if apiPort == "" {
		apiPort = "8080" // Valor por defecto
	}

	server := &http.Server{
		Addr:         ":" + apiPort,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Printf("🚀 Servidor escuchando en el puerto %s", apiPort)
	log.Printf("📘 Documentación de la API disponible en http://localhost:%s/swagger/index.html", apiPort)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("FATAL: No se pudo iniciar el servidor: %v", err)
	}
}

// setupRouter configura el router de Gin y registra todas las rutas.
func setupRouter(containerHandler *container.Handler) *gin.Engine {
	// gin.SetMode(gin.ReleaseMode) // Descomentar para producción
	router := gin.Default()

	// Middleware de logging y recuperación de panics (ya incluido en gin.Default())
	// Se podrían añadir otros middlewares aquí (CORS, etc.)

	// Ruta de Health Check simple
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "pong"})
	})

	// Grupo de rutas para la v1 de la API
	v1 := router.Group("/api/v1")
	{
		// Registramos las rutas del módulo de contenedores
		containerHandler.RegisterRoutes(v1)
	}

	// Ruta para la documentación de Swagger
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return router
}
