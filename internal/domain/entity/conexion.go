package entity

import (
	"time"

	"DBGS_SOBERANO_BACKEND/internal/domain"
)

// Conexion representa una conexión registrada a una base de datos externa.
// El password se almacena cifrado y nunca se expone hacia el cliente.
type Conexion struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Engine        string    `json:"engine"` // postgresql | mysql
	Host          string    `json:"host"`
	Port          int       `json:"port"`
	User          string    `json:"user"`
	PasswordHash  string    `json:"-"` // cifrado, no se serializa
	Database      string    `json:"database"`
	SSLMode       string    `json:"ssl_mode"`
	ReadOnly      bool      `json:"read_only"`
	CreatedAt     time.Time `json:"created_at"`
	CreatedBy     string    `json:"created_by"`
}

// EsValido valida los campos obligatorios de la conexión.
func (c *Conexion) EsValido() error {
	if c.Name == "" || c.Host == "" || c.Database == "" || c.User == "" {
		return domain.RequiredError("conexion")
	}
	if c.Engine != "postgresql" && c.Engine != "mysql" {
		return domain.InvalidValueError("engine", []string{"postgresql", "mysql"})
	}
	if c.Port <= 0 {
		return domain.RangeError("port", 1, 65535)
	}
	return nil
}

// ApiPublicada representa una API pública de solo lectura sobre una tabla externa.
type ApiPublicada struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	Slug           string    `json:"slug"`
	ConnectionID   string    `json:"connection_id"`
	ConnectionName string    `json:"connection_name"`
	Schema         string    `json:"schema"`
	Table          string    `json:"table"`
	MaxRows        int       `json:"max_rows"`
	Active         bool      `json:"active"`
	APIKey         string    `json:"api_key"`
	Endpoint       string    `json:"endpoint"`
	CreatedAt      time.Time `json:"created_at"`
	CreatedBy      string    `json:"created_by"`
}

// EsValido valida los campos obligatorios de la API publicada.
func (a *ApiPublicada) EsValido() error {
	if a.Name == "" || a.Slug == "" || a.ConnectionID == "" || a.Schema == "" || a.Table == "" {
		return domain.RequiredError("api")
	}
	return nil
}
