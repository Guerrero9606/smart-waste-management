package domain

import (
	"time"
)

// Status define el tipo para el estado de llenado de un contenedor.
// Usar un tipo personalizado mejora la legibilidad y la seguridad de tipos.
type Status string

// Constantes que definen los posibles estados de un contenedor.
// Corresponden al tipo ENUM 'container_status' en la base de datos.
const (
	StatusLow    Status = "low"
	StatusMedium Status = "medium"
	StatusHigh   Status = "high"
)

// Point representa una coordenada geográfica.
type Point struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// Container representa la entidad principal de nuestro dominio.
// Contiene la información estática y el estado actual de un contenedor de basura.
type Container struct {
	ID             string    `json:"id"`
	Location       Point     `json:"location"`
	CapacityLiters int       `json:"capacity_liters"`
	CurrentStatus  Status    `json:"status,omitempty"` // omitempty porque no se establece al crear
	LastFillLevel  int       `json:"last_fill_level,omitempty"`
	LastUpdatedAt  time.Time `json:"last_updated,omitempty"`

	// --- CAMPOS ACTUALIZADOS ---
	// Estos campos son gestionados por la base de datos y son cruciales para el tracking.
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Reading representa una única lectura del sensor de un contenedor.
// Es un evento inmutable que ocurrió en un momento específico.
type Reading struct {
	ContainerID string    `json:"container_id"`
	FillLevel   int       `json:"fill_level"`
	Timestamp   time.Time `json:"timestamp"`
}

// === Lógica de Negocio Pura ===

// CalculateStatus determina el estado del contenedor ('low', 'medium', 'high')
// basándose en su nivel de llenado.
func CalculateStatus(fillLevel int) Status {
	if fillLevel >= 80 {
		return StatusHigh
	}
	if fillLevel >= 40 {
		return StatusMedium
	}
	return StatusLow
}

// IsValid comprueba si los datos de una nueva lectura son válidos.
func (r *Reading) IsValid() bool {
	if r.ContainerID == "" {
		return false
	}
	if r.FillLevel < 0 || r.FillLevel > 100 {
		return false
	}
	if r.Timestamp.IsZero() {
		return false
	}
	return true
}
