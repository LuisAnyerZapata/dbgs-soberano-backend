package entity

import "time"

// Institucion representa una entidad gubernamental adscrita o integrante de la red DBGS
type Institucion struct {
    ID                  string    `json:"id"`
    CodigoInstitucional string    `json:"codigo_institucional"`
    NombreInstitucion   string    `json:"nombre_institucion"`
    Siglas              string    `json:"siglas"`
    Estatus             bool      `json:"estatus"`
    CreatedAt           time.Time `json:"created_at"`
    CreatedBy           string    `json:"created_by"`
    UpdatedAt           time.Time `json:"updated_at"`
    UpdatedBy           string    `json:"updated_by"`
}