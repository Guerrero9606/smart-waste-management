package container

import (
	"context"
	"fmt"
	"math"
	"smart-waste-management/internal/domain"
)

// Service define la interfaz para la lógica de negocio relacionada con los contenedores.
// Esta abstracción permite que los handlers dependan de la interfaz, no de la implementación concreta.
type Service interface {
	// ProcessNewReading valida y procesa una nueva lectura de un sensor.
	ProcessNewReading(ctx context.Context, reading domain.Reading) error
	// GetAllContainers obtiene todos los contenedores para su visualización.
	GetAllContainers(ctx context.Context) ([]domain.Container, error)
	// GenerateRoute crea una ruta de recogida optimizada.
	GenerateRoute(ctx context.Context, startPoint domain.Point, statuses []domain.Status) ([]domain.Container, error)

	CreateContainer(ctx context.Context, container domain.Container) (domain.Container, error)
	GetContainerByID(ctx context.Context, id string) (domain.Container, error)
	UpdateContainer(ctx context.Context, container domain.Container) error
	DeleteContainer(ctx context.Context, id string) error
	GetReadingsForContainer(ctx context.Context, id string, limit int) ([]domain.Reading, error)
}

// service es la implementación concreta de la interfaz Service.
type service struct {
	repo Repository // Depende de la interfaz del Repositorio, no de su implementación.
}

// NewService crea una nueva instancia del servicio.
// Recibe el repositorio como una dependencia (Inyección de Dependencias).
func NewService(repo Repository) Service {
	return &service{
		repo: repo,
	}
}

// ProcessNewReading contiene la lógica de negocio para procesar una nueva lectura.
func (s *service) ProcessNewReading(ctx context.Context, reading domain.Reading) error {
	// 1. Validación de negocio.
	// La capa de servicio es el lugar ideal para este tipo de reglas.
	if !reading.IsValid() {
		return fmt.Errorf("la lectura proporcionada no es válida: %+v", reading)
	}

	// 2. Aquí se podrían añadir más lógicas de negocio complejas.
	// Por ejemplo:
	// - Comprobar si el `container_id` existe antes de intentar guardar (aunque la FK de la BBDD ya lo hace).
	// - Enviar una notificación (ej. un evento a Kafka o RabbitMQ) si el `fill_level` supera el 95%.
	// - Comprobar si la lectura es más antigua que la última registrada y, en ese caso, ignorarla.

	fmt.Printf("Procesando nueva lectura para el contenedor %s con nivel %d%%\n", reading.ContainerID, reading.FillLevel)

	// 3. Delegar la persistencia al repositorio.
	// El servicio no sabe cómo se guarda, solo que debe guardarse.
	err := s.repo.SaveReading(ctx, reading)
	if err != nil {
		// Envolvemos el error del repositorio para dar más contexto.
		return fmt.Errorf("error al guardar la lectura en el repositorio: %w", err)
	}

	return nil
}

// GetAllContainers simplemente delega la llamada al repositorio.
// En un caso más complejo, podría enriquecer los datos antes de devolverlos.
func (s *service) GetAllContainers(ctx context.Context) ([]domain.Container, error) {
	fmt.Println("Obteniendo todos los contenedores desde el servicio.")

	containers, err := s.repo.FindAllContainers(ctx)
	if err != nil {
		// Envolvemos el error del repositorio.
		return nil, fmt.Errorf("error al obtener los contenedores desde el repositorio: %w", err)
	}

	// Lógica de negocio adicional podría ir aquí.
	// Por ejemplo, si el frontend necesita un campo extra que no está en la BBDD,
	// se podría calcular y añadir aquí.

	return containers, nil
}

func (s *service) GenerateRoute(ctx context.Context, startPoint domain.Point, statuses []domain.Status) ([]domain.Container, error) {
	// 1. Obtener todos los contenedores que cumplen con el criterio desde el repositorio.
	containersToVisit, err := s.repo.FindContainersByStatus(ctx, statuses)
	if err != nil {
		return nil, fmt.Errorf("no se pudieron obtener los contenedores para la ruta: %w", err)
	}

	if len(containersToVisit) == 0 {
		return []domain.Container{}, nil // No hay contenedores que visitar, devolvemos una ruta vacía.
	}

	// 2. Aplicar el algoritmo de optimización (Vecino más cercano).
	var route []domain.Container
	currentPoint := startPoint

	for len(containersToVisit) > 0 {
		nearestIndex := -1
		minDistance := math.MaxFloat64

		// Encontrar el contenedor más cercano al punto actual.
		for i, container := range containersToVisit {
			dist := haversineDistance(currentPoint, container.Location)
			if dist < minDistance {
				minDistance = dist
				nearestIndex = i
			}
		}

		// Añadir el contenedor más cercano a la ruta.
		nearestContainer := containersToVisit[nearestIndex]
		route = append(route, nearestContainer)

		// Actualizar el punto actual para la siguiente iteración.
		currentPoint = nearestContainer.Location

		// Eliminar el contenedor visitado de la lista de pendientes.
		containersToVisit = append(containersToVisit[:nearestIndex], containersToVisit[nearestIndex+1:]...)
	}

	return route, nil
}

func (s *service) CreateContainer(ctx context.Context, container domain.Container) (domain.Container, error) {
	// Aquí podría ir la validación de negocio, por ejemplo, comprobar si la capacidad es válida.
	return s.repo.CreateContainer(ctx, container)
}

func (s *service) GetContainerByID(ctx context.Context, id string) (domain.Container, error) {
	return s.repo.FindContainerByID(ctx, id)
}

func (s *service) UpdateContainer(ctx context.Context, container domain.Container) error {
	return s.repo.UpdateContainer(ctx, container)
}

func (s *service) DeleteContainer(ctx context.Context, id string) error {
	return s.repo.DeleteContainer(ctx, id)
}

func (s *service) GetReadingsForContainer(ctx context.Context, id string, limit int) ([]domain.Reading, error) {
	if limit <= 0 || limit > 100 { // Ponemos un límite por defecto y máximo
		limit = 50
	}
	return s.repo.FindReadingsByContainerID(ctx, id, limit)
}

// haversineDistance calcula la distancia en kilómetros entre dos puntos geográficos.
// Es una función de utilidad que podemos añadir al final del fichero.
func haversineDistance(p1, p2 domain.Point) float64 {
	const R = 6371 // Radio de la Tierra en kilómetros
	lat1Rad := p1.Latitude * math.Pi / 180
	lon1Rad := p1.Longitude * math.Pi / 180
	lat2Rad := p2.Latitude * math.Pi / 180
	lon2Rad := p2.Longitude * math.Pi / 180

	dLon := lon2Rad - lon1Rad
	dLat := lat2Rad - lat1Rad

	a := math.Pow(math.Sin(dLat/2), 2) + math.Cos(lat1Rad)*math.Cos(lat2Rad)*math.Pow(math.Sin(dLon/2), 2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return R * c
}
