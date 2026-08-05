package entity

import (
	"time"
)

// Rol representa la agrupación de permisos por función
type Rol struct {
	ID          string    `json:"id"`
	Nombre      string    `json:"nombre"` // Ej: "ADMINISTRADOR", "AUDITOR", "SERVICIO"
	Descripcion string    `json:"descripcion"`
	Permisos    []string  `json:"permisos"` // Lista de permisos asociados
	Estado      bool      `json:"estado"`
}

// Usuario representa la identidad humana o cuenta técnica autorizada
type Usuario struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	RolID     string    `json:"rol_id"`
	Rol       Rol       `json:"rol,omitempty"`
	EsTecnico bool      `json:"es_tecnico"` // Separación entre cuentas humanas y de servicio
	Estado    bool      `json:"estado"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (u *Usuario) EstaActivo() bool {
	return u.Estado
}