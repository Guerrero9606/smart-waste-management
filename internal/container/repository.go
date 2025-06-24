package container

import (
	"context"
	"fmt"
	"smart-waste-management/internal/domain"
	"smart-waste-management/internal/platform/database"
	"time"

	"github.com/jackc/pgx/v5"
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
	// FindContainersByStatus busca contenedores por su estado actual y devuelve sus IDs y ubicaciones.
	FindContainersByStatus(ctx context.Context, statuses []domain.Status) ([]domain.Container, error)

	CreateContainer(ctx context.Context, container domain.Container) (domain.Container, error)
	UpdateContainer(ctx context.Context, container domain.Container) error
	DeleteContainer(ctx context.Context, id string) error
	FindReadingsByContainerID(ctx context.Context, id string, limit int) ([]domain.Reading, error)
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
               capacity_liters, current_status, last_fill_level, last_updated_at,
               created_at, updated_at
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
		var lastUpdatedAt time.Time

		err := rows.Scan(
			&c.ID, &c.Location.Latitude, &c.Location.Longitude,
			&c.CapacityLiters, &c.CurrentStatus, &c.LastFillLevel,
			&lastUpdatedAt, &c.CreatedAt, &c.UpdatedAt, // Añadimos los nuevos campos al Scan
		)
		if err != nil {
			return nil, fmt.Errorf("error al escanear la fila del contenedor: %w", err)
		}
		c.LastUpdatedAt = lastUpdatedAt
		containers = append(containers, c)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("error durante la iteración de las filas: %w", rows.Err())
	}

	return containers, nil
}

func (r *postgresRepository) FindContainersByStatus(ctx context.Context, statuses []domain.Status) ([]domain.Container, error) {
	// --- INICIO DE LA MODIFICACIÓN ---
	// Creamos un slice de strings vacío para la conversión explícita.
	stringStatuses := make([]string, len(statuses))
	// Iteramos sobre nuestro slice de domain.Status y lo convertimos a un slice de string.
	for i, s := range statuses {
		stringStatuses[i] = string(s)
	}
	// --- FIN DE LA MODIFICACIÓN ---

	// AÑADE ESTA LÍNEA PARA DEPURAR
	fmt.Println(">>> DEBUG: Ejecutando consulta con statuses convertidos:", stringStatuses)

	query := `
        SELECT id, ST_Y(location::geometry) as latitude, ST_X(location::geometry) as longitude
        FROM containers
        WHERE current_status = ANY($1)
        ORDER BY id; -- Ordenar para tener un resultado consistente
    `

	// Le pasamos el nuevo slice de strings a la consulta.
	rows, err := r.db.Query(ctx, query, stringStatuses)
	if err != nil {
		return nil, fmt.Errorf("error al consultar contenedores por estado: %w", err)
	}
	defer rows.Close()

	var containers []domain.Container
	for rows.Next() {
		var c domain.Container
		err := rows.Scan(&c.ID, &c.Location.Latitude, &c.Location.Longitude)
		if err != nil {
			return nil, fmt.Errorf("error al escanear contenedor por estado: %w", err)
		}
		containers = append(containers, c)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("error durante la iteración de filas por estado: %w", rows.Err())
	}

	return containers, nil
}

func (r *postgresRepository) CreateContainer(ctx context.Context, container domain.Container) (domain.Container, error) {
	query := `
        INSERT INTO containers (location, capacity_liters)
        VALUES (ST_SetSRID(ST_MakePoint($1, $2), 4326), $3)
        RETURNING id, created_at, updated_at`

	err := r.db.QueryRow(ctx, query, container.Location.Longitude, container.Location.Latitude, container.CapacityLiters).Scan(
		&container.ID,
		&container.CreatedAt, // Asumiendo que has añadido CreatedAt y UpdatedAt a tu struct de dominio
		&container.UpdatedAt,
	)

	if err != nil {
		return domain.Container{}, fmt.Errorf("error al crear el contenedor: %w", err)
	}
	return container, nil
}

func (r *postgresRepository) FindContainerByID(ctx context.Context, id string) (domain.Container, error) {
	query := `
        SELECT id, ST_Y(location::geometry) as latitude, ST_X(location::geometry) as longitude,
               capacity_liters, current_status, last_fill_level, last_updated_at,
               created_at, updated_at
        FROM containers
        WHERE id = $1`

	var c domain.Container
	var lastUpdatedAt time.Time

	err := r.db.QueryRow(ctx, query, id).Scan(
		&c.ID, &c.Location.Latitude, &c.Location.Longitude,
		&c.CapacityLiters, &c.CurrentStatus, &c.LastFillLevel,
		&lastUpdatedAt, &c.CreatedAt, &c.UpdatedAt, // Añadimos los nuevos campos al Scan
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return domain.Container{}, fmt.Errorf("contenedor no encontrado") // Error específico para "no encontrado"
		}
		return domain.Container{}, fmt.Errorf("error al buscar contenedor por ID: %w", err)
	}
	c.LastUpdatedAt = lastUpdatedAt
	return c, nil
}

func (r *postgresRepository) UpdateContainer(ctx context.Context, container domain.Container) error {
	query := `
        UPDATE containers
        SET location = ST_SetSRID(ST_MakePoint($1, $2), 4326), capacity_liters = $3, updated_at = NOW()
        WHERE id = $4`

	_, err := r.db.Exec(ctx, query, container.Location.Longitude, container.Location.Latitude, container.CapacityLiters, container.ID)
	return err
}

func (r *postgresRepository) DeleteContainer(ctx context.Context, id string) error {
	query := `DELETE FROM containers WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *postgresRepository) FindReadingsByContainerID(ctx context.Context, id string, limit int) ([]domain.Reading, error) {
	query := `
        SELECT container_id, fill_level, recorded_at
        FROM readings
        WHERE container_id = $1
        ORDER BY recorded_at DESC
        LIMIT $2`

	rows, err := r.db.Query(ctx, query, id, limit)
	if err != nil {
		return nil, fmt.Errorf("error al consultar lecturas: %w", err)
	}
	defer rows.Close()

	var readings []domain.Reading
	for rows.Next() {
		var r domain.Reading
		if err := rows.Scan(&r.ContainerID, &r.FillLevel, &r.Timestamp); err != nil {
			return nil, fmt.Errorf("error al escanear lectura: %w", err)
		}
		readings = append(readings, r)
	}
	return readings, rows.Err()
}
