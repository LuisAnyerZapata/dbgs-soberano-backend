package port

import (
    "context"
    "DBGS_SOBERANO_BACKEND/internal/domain/entity"
)

type CrearColeccionInput struct {
    Nombre       string                  `json:"nombre"`
    Descripcion string                  `json:"descripcion"`
    Campos       []entity.CampoDinamico `json:"campos"`
    InstitucionID string                  `json:"institucion_id"`
}

type ListarColeccionesOutput struct {
    Colecciones []entity.ColeccionRegistro `json:"colecciones"`
    Total       int64                    `json:"total"`
}

type ColeccionPort interface {
    CrearColeccion(ctx context.Context, input CrearColeccionInput) (*entity.ColeccionRegistro, error)
    ListarColecciones(ctx context.Context, limite, offset int) (*ListarColeccionesOutput, error) // NUEVO
}