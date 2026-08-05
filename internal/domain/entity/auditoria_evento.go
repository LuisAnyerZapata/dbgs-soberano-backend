package entity

import (
	"time"
)

// AuditoriaEvento registra las operaciones sensibles y accesos al sistema
type AuditoriaEvento struct {
	ID            string    `json:"id"`
	UsuarioID     string    `json:"usuario_id"`     // Usuario o cuenta técnica responsable
	Username      string    `json:"username"`       // Nombre de usuario al momento del evento
	Operacion     string    `json:"operacion"`      // Acción realizada (ej: "CONSULTA_CATALOGO", "CREAR_SCHEMA")
	Recurso       string    `json:"recurso"`        // Entidad o endpoint accedido
	Detalles      string    `json:"detalles"`       // Información contextual de la operación
	Resultado     string    `json:"resultado"`      // "EXITO", "DENEGADO", "ERROR"
	IPOrigen      string    `json:"ip_origen"`      // Dirección IP de origen
	FechaCreacion time.Time `json:"fecha_creacion"` // Timestamp preciso del evento
}
