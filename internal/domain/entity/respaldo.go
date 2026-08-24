package entity

import "time"

// Estados válidos del ciclo de vida de una operación de respaldo/restauración.
const (
	EstadoEnProgreso = "EN_PROGRESO"
	EstadoCompletado = "COMPLETADO"
	EstadoFallido    = "FALLIDO"
)

// RespaldoOperacion representa un respaldo generado por el sistema.
// FechaFinalizacion es cero mientras la operación siga en progreso.
type RespaldoOperacion struct {
	ID                string    `json:"id"`
	Tipo              string    `json:"tipo"`
	Estado            string    `json:"estado"`
	RutaArchivo       string    `json:"ruta_archivo"`
	TamanoBytes       int64     `json:"tamano_bytes"`
	Detalles          string    `json:"detalles"`
	FechaCreacion     time.Time `json:"fecha_creacion"`
	FechaFinalizacion time.Time `json:"fecha_finalizacion"`
	RetencionDias     int       `json:"retencion_dias"`
	UsuarioCreador    string    `json:"usuario_creador"`
}

// Restauracion representa una restauración validada del sistema.
type Restauracion struct {
	ID            string    `json:"id"`
	BackupID      string    `json:"backup_id"`
	Usuario       string    `json:"usuario"`
	Estado        string    `json:"estado"`
	Validado      bool      `json:"validado"`
	FechaCreacion time.Time `json:"fecha_creacion"`
	Observaciones string    `json:"observaciones"`
}

// LogOperativo registra eventos del sistema para observabilidad.
type LogOperativo struct {
	ID            string    `json:"id"`
	Nivel         string    `json:"nivel"`
	Modulo        string    `json:"modulo"`
	Mensaje       string    `json:"mensaje"`
	FechaCreacion time.Time `json:"fecha_creacion"`
}

// MetricaSistema representa una métrica básica del sistema.
type MetricaSistema struct {
	ID            string    `json:"id"`
	Nombre        string    `json:"nombre"`
	Valor         float64   `json:"valor"`
	Unidad        string    `json:"unidad"`
	FechaCreacion time.Time `json:"fecha_creacion"`
}

// HealthCheck representa el resultado de un chequeo de salud.
type HealthCheck struct {
	ID            string    `json:"id"`
	Componente    string    `json:"componente"`
	Estado        string    `json:"estado"`
	Mensaje       string    `json:"mensaje"`
	FechaCreacion time.Time `json:"fecha_creacion"`
}
