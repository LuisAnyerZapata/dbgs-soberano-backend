package entity

import "time"

// ClienteIntegracion representa un consumidor externo autorizado para interoperar con el backend.
type ClienteIntegracion struct {
    ID              string    `json:"id"`
    Nombre          string    `json:"nombre"`
    Tipo            string    `json:"tipo"`
    VersionContrato string    `json:"version_contrato"`
    TokenHash       string    `json:"-"` // SECRETO: El hash SHA-256. El tag `json:"-"` evita que se filtre en respuestas HTTP por error.
    TokenPlano      string    `json:"token_plano,omitempty"` // TEMPORAL: Solo se popula al momento de crear el cliente, para mostrárselo a la UI una única vez.
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