package entity

import (
    "encoding/json"
    "time"
    "database/sql"
)

// FieldType define los tipos de datos permitidos para los campos dinámicos.
// Se utiliza un enum seguro para evitar inyecciones de tipos no deseados.
type FieldType string

const (
    FieldTypeString  FieldType = "STRING"
    FieldTypeText    FieldType = "TEXT"
    FieldTypeInt     FieldType = "INTEGER"
    FieldTypeFloat   FieldType = "FLOAT"
    FieldTypeBoolean FieldType = "BOOLEAN"
    FieldTypeUUID    FieldType = "UUID"
    FieldTypeJSON    FieldType = "JSONB"
    FieldTypeDate    FieldType = "TIMESTAMPTZ"
)

// CampoDinamico representa la definición de una columna para la nueva tabla
type CampoDinamico struct {
    Nombre     string    `json:"nombre"`
    Tipo       FieldType `json:"tipo"`
    Nulo       bool      `json:"nulo"`
    Unico      bool      `json:"unico"`
    EsPK       bool      `json:"es_pk"` // Opcional, por si quieren definir su propio PK (aunque el sistema forzará un UUID por defecto)
    Descripcion string    `json:"descripcion"`
}

// ColeccionDefinicion es el DTO que recibirá el Use Case para pedir la creación de una tabla
type ColeccionDefinicion struct {
    Nombre       string           `json:"nombre"`
    Descripcion string           `json:"descripcion"`
    Campos       []CampoDinamico  `json:"campos"`
    InstitucionID string           `json:"institucion_id"` // Preparado para el RLS del futuro
}

// ColeccionRegistro representa la fila en la tabla dbgs_schema.colecciones_dinamicas
type ColeccionRegistro struct {
    ID             string          `json:"id"`
    NombreLogico   string          `json:"nombre_logico"`
    NombreFisico   string          `json:"nombre_fisico"`
    Descripcion    string          `json:"descripcion"`
    InstitucionID  sql.NullString  `json:"institucion_id"`
    EstructuraJSON json.RawMessage `json:"estructura"` // Almacena el array de CampoDinamico como JSON
    EstaActiva     bool            `json:"esta_activa"`
    CreatedAt      time.Time       `json:"created_at"`
    CreatedBy      string          `json:"created_by"`
}