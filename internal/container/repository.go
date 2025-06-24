package container

import (
	"context"
	"fmt"
	"smart-waste-management/internal/domain"
	"smart-waste-management/internal/platform/database"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository define la interfaz para las operaciones de persistencia de contenedores.
// Usar una interfaz nos permitirá 'mockear' el repositorio fácilmente para las pruebas unitarias del servicio.
type Repository interface {
	// SaveReading guarda una nueva lectura y actualiza el estado del contenedor correspondiente.
	SaveReading(ctx context.Context, reading domain.Reading) error
	// FindAllContainers devuelve todos los contenedores con su estado actual.
	FindAllContainers(ctx context.Context) ([]domain.Container, error)
	// FindContainerByID busca un único contenedor por su ID.
	FindContainerByID(ctx context.Context, id string) (domain.Container, error)
}

// postgresRepository es la implementación concreta de la interfaz Repository para PostgreSQL.
type postgresRepository struct {
	db *pgxpool.Pool
}

// NewPostgresRepository crea una nueva instancia del repositorio.
// Recibe el pool de conexiones como una dependencia.
func NewPostgresRepository(db *database.DB) Repository {
	return &postgresRepository{
		db: db.Pool,
	}
}

// SaveReading implementa la lógica para guardar una lectura en la base de datos.
// Se ejecuta dentro de una transacción para garantizar la consistencia de los datos.
func (r *postgresRepository) SaveReading(ctx context.Context, reading domain.Reading) error {
	// Calculamos el nuevo estado basado en la lógica de dominio.
	newStatus := domain.CalculateStatus(reading.FillLevel)

	// Iniciamos una transacción. Si cualquiera de las dos operaciones (INSERT o UPDATE) falla,
	// se hará un rollback automático de ambas, manteniendo la base de datos consistente.
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("no se pudo iniciar la transacción: %w", err)
	}
	defer tx.Rollback(ctx) // Defer Rollback es un patrón seguro. Si Commit() tiene éxito, no hace nada.

	// 1. Insertamos la nueva lectura en la tabla 'readings'.
	insertReadingSQL := `
        INSERT INTO readings (container_id, fill_level, recorded_at)
        VALUES ($1, $2, $3)`
	_, err = tx.Exec(ctx, insertReadingSQL, reading.ContainerID, reading.FillLevel, reading.Timestamp)
	if err != nil {
		return fmt.Errorf("error al insertar la lectura: %w", err)
	}

	// 2. Actualizamos el estado denormalizado en la tabla 'containers'.
	updateContainerSQL := `
        UPDATE containers
        SET current_status = $1, last_fill_level = $2, last_updated_at = $3, updated_at = NOW()
        WHERE id = $4`
	_, err = tx.Exec(ctx, updateContainerSQL, newStatus, reading.FillLevel, reading.Timestamp, reading.ContainerID)
	if err != nil {
		return fmt.Errorf("error al actualizar el contenedor: %w", err)
	}

	// Si ambas operaciones fueron exitosas, hacemos commit de la transacción.
	return tx.Commit(ctx)
}

// FindAllContainers recupera todos los contenedores de la base de datos.
func (r *postgresRepository) FindAllContainers(ctx context.Context) ([]domain.Container, error) {
	query := `
        SELECT id, ST_Y(location::geometry) as latitude, ST_X(location::geometry) as longitude,
               capacity_liters, current_status, last_fill_level, last_updated_at
        FROM containers
        ORDER BY created_at DESC`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error al consultar los contenedores: %w", err)
	}
	defer rows.Close()

	var containers []domain.Container
	for rows.Next() {
		var c domain.Container
		var lastUpdatedAt time.Time // pgx necesita un puntero a time.Time para manejar valores NULL

		err := rows.Scan(
			&c.ID,
			&c.Location.Latitude,
			&c.Location.Longitude,
			&c.CapacityLiters,
			&c.CurrentStatus,
			&c.LastFillLevel,
			&lastUpdatedAt, // Escanea a la variable temporal
		)
		if err != nil {
			return nil, fmt.Errorf("error al escanear la fila del contenedor: %w", err)
		}
		c.LastUpdatedAt = lastUpdatedAt // Asigna el valor
		containers = append(containers, c)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("error durante la iteración de las filas: %w", rows.Err())
	}

	return containers, nil
}

// FindContainerByID busca un único contenedor.
// (Implementación similar a FindAllContainers pero con un WHERE id = $1)
func (r *postgresRepository) FindContainerByID(ctx context.Context, id string) (domain.Container, error) {
	// Implementar la lógica para buscar un solo contenedor si es necesario.
	// Por ahora, lo dejamos como un placeholder.
	return domain.Container{}, fmt.Errorf("método no implementado")
}
