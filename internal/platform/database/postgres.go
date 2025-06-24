package database

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DB es un pool de conexiones a la base de datos PostgreSQL.
// Usamos pgxpool porque gestiona un conjunto de conexiones de forma eficiente,
// lo cual es crucial para una aplicación web con peticiones concurrentes.
type DB struct {
	Pool *pgxpool.Pool
}

// NewDBConnection crea y configura una nueva conexión a la base de datos
// a partir de las variables de entorno y devuelve un puntero al pool de conexiones.
func NewDBConnection() (*DB, error) {
	// DSN (Data Source Name) es la cadena de conexión.
	// Leemos los parámetros desde las variables de entorno para mayor seguridad y flexibilidad.
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_SSLMODE"),
	)

	// pgxpool.ParseConfig es más robusto que pasar directamente la DSN,
	// ya que permite una configuración más detallada.
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("no se pudo parsear la configuración de la BBDD: %w", err)
	}

	// === Configuración del Pool de Conexiones ===
	// Estos valores son un buen punto de partida para la mayoría de aplicaciones.
	// MaxConns: Número máximo de conexiones que el pool puede tener.
	config.MaxConns = 10
	// MinConns: Número mínimo de conexiones que el pool mantendrá abiertas.
	config.MinConns = 2
	// MaxConnLifetime: Tiempo máximo que una conexión puede ser reutilizada.
	config.MaxConnLifetime = time.Hour
	// MaxConnIdleTime: Tiempo máximo que una conexión puede estar inactiva en el pool.
	config.MaxConnIdleTime = 30 * time.Minute
	// HealthCheckPeriod: Frecuencia con la que se comprueba la salud de las conexiones.
	config.HealthCheckPeriod = 1 * time.Minute
	// ConnectTimeout: Tiempo máximo de espera para establecer una conexión.
	config.ConnectTimeout = 5 * time.Second

	// Creamos un contexto con timeout para el intento de conexión inicial.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("no se pudo crear el pool de conexiones a la BBDD: %w", err)
	}

	// Hacemos un ping para verificar que la conexión es exitosa.
	if err := pool.Ping(ctx); err != nil {
		pool.Close() // Si el ping falla, cerramos el pool para liberar recursos.
		return nil, fmt.Errorf("no se pudo hacer ping a la BBDD: %w", err)
	}

	fmt.Println("¡Conexión a la base de datos PostgreSQL establecida con éxito!")

	return &DB{Pool: pool}, nil
}

// Close cierra todas las conexiones en el pool.
// Se debe llamar a esta función cuando la aplicación se está cerrando (graceful shutdown).
func (db *DB) Close() {
	fmt.Println("Cerrando conexiones de la base de datos...")
	db.Pool.Close()
}
