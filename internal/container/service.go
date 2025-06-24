package container

import (
	"context"
	"fmt"
	"smart-waste-management/internal/domain"
)

// Service define la interfaz para la lógica de negocio relacionada con los contenedores.
// Esta abstracción permite que los handlers dependan de la interfaz, no de la implementación concreta.
type Service interface {
	// ProcessNewReading valida y procesa una nueva lectura de un sensor.
	ProcessNewReading(ctx context.Context, reading domain.Reading) error
	// GetAllContainers obtiene todos los contenedores para su visualización.
	GetAllContainers(ctx context.Context) ([]domain.Container, error)
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
