// internal/platform/database/postgres.go

package database

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	Pool *pgxpool.Pool
}

func NewDBConnection() (*DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_SSLMODE"),
	)

	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("no se pudo parsear la configuración de la BBDD: %w", err)
	}

	// --- INICIO DE LA MODIFICACIÓN: CORRECCIÓN DEL REGISTRO DE TIPOS ---
	// La función AfterConnect sigue siendo la forma más idiomática y segura de asegurar
	// que CADA conexión del pool conozca nuestros tipos personalizados.
	config.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		// 1. Cargamos la definición del tipo 'container_status' desde la base de datos.
		dataType, err := conn.LoadType(ctx, "container_status")
		if err != nil {
			return fmt.Errorf("no se pudo cargar el tipo 'container_status' desde la BBDD: %w", err)
		}
		// 2. Registramos este tipo en el mapa de tipos de la conexión actual.
		conn.TypeMap().RegisterType(dataType)
		return nil
	}
	// --- FIN DE LA MODIFICACIÓN ---

	config.MaxConns = 10
	config.MinConns = 2
	config.MaxConnLifetime = time.Hour
	config.MaxConnIdleTime = 30 * time.Minute
	config.HealthCheckPeriod = 1 * time.Minute

	connectCtx, cancelConnect := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelConnect()

	pool, err := pgxpool.NewWithConfig(connectCtx, config)
	if err != nil {
		return nil, fmt.Errorf("no se pudo crear el pool de conexiones a la BBDD: %w", err)
	}

	pingCtx, cancelPing := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancelPing()

	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("no se pudo hacer ping a la BBDD: %w", err)
	}

	fmt.Println("¡Conexión a la base de datos PostgreSQL establecida con éxito!")

	return &DB{Pool: pool}, nil
}

func (db *DB) Close() {
	fmt.Println("Cerrando conexiones de la base de datos...")
	db.Pool.Close()
}
