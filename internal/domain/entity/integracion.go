package entity

import "time"

// ClienteIntegracion representa un consumidor externo autorizado para interoperar con el backend.
type ClienteIntegracion struct {
	ID              string    `json:"id"`
	Nombre          string    `json:"nombre"`
	Tipo            string    `json:"tipo"`
	VersionContrato string    `json:"version_contrato"`
	Token           string    `json:"token"`
	Estado          bool      `json:"estado"`
	Scopes          []string  `json:"scopes"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// SolicitudIntegracion registra un intento de consumo de un servicio por parte de un cliente.
type SolicitudIntegracion struct {
	ID         string    `json:"id"`
	ClienteID  string    `json:"cliente_id"`
	Metodo     string    `json:"metodo"`
	Recurso    string    `json:"recurso"`
	Estado     string    `json:"estado"`
	Fecha      time.Time `json:"fecha"`
	Detalles   string    `json:"detalles,omitempty"`
}
