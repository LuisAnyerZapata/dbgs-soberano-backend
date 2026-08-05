package entity

import (
	"time"

	"DBGS_SOBERANO_BACKEND/internal/domain"
)

// FuenteDato representa el origen de la información.
type FuenteDato struct {
	ID          string    `json:"id"`
	Nombre      string    `json:"nombre"`
	Descripcion string    `json:"descripcion"`
	Estado      bool      `json:"estado"`
	CreatedAt   time.Time `json:"created_at"`
	CreatedBy   string    `json:"created_by"`
}

// ConjuntoDato representa un Dataset dentro del dominio.
type ConjuntoDato struct {
	ID               string    `json:"id"`
	FuenteDatoID     string    `json:"fuente_dato_id"`
	Nombre           string    `json:"nombre"`
	Proposito        string    `json:"proposito"`
	PropietarioDato  string    `json:"propietario_dato"`
	Clasificacion    string    `json:"clasificacion"`
	Estado           bool      `json:"estado"`
	CreatedAt        time.Time `json:"created_at"`
	CreatedBy        string    `json:"created_by"`
	UpdatedAt        time.Time `json:"updated_at"`
	UpdatedBy        string    `json:"updated_by"`
}

// EsValido verifica la validez del dataset.
func (cd *ConjuntoDato) EsValido() error {
	if cd.Nombre == "" || cd.FuenteDatoID == "" {
		return domain.ErrDatosInvalidos
	}
	return nil
}