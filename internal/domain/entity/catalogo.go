package entity

import "time"

type Catalogo struct {
	ID          string    `json:"id"`
	Codigo      string    `json:"codigo"`
	Nombre      string    `json:"nombre"`
	Descripcion string    `json:"descripcion"`
	Estado      bool      `json:"estado"`
	CreatedAt   time.Time `json:"created_at"`
	CreatedBy   string    `json:"created_by"`
	UpdatedAt   time.Time `json:"updated_at"`
	UpdatedBy   string    `json:"updated_by"`
}

// Inactivar cambia el estado del catálogo a inactivo y actualiza la marca de tiempo.
func (c *Catalogo) Inactivar() {
	c.Estado = false
	c.UpdatedAt = time.Now()
}
