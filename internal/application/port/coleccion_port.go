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

type ActualizarColeccionInput struct {
    Nombre   string                  `json:"nombre"`
    Campos   []entity.CampoDinamico `json:"campos"`
}

type ActualizarColeccionOutput struct {
    ID             string `json:"id"`
    NombreLogico   string `json:"nombre_logico"`
    CamposAgregados int   `json:"campos_agregados"`
}

type EliminarColeccionInput struct {
    Nombre   string `json:"nombre"`
    Confirmar bool  `json:"confirmar"`
}

type EliminarColeccionOutput struct {
    NombreLogico   string `json:"nombre_logico"`
    AccionRealizada string `json:"accion_realizada"`
}

type ColeccionPort interface {
    CrearColeccion(ctx context.Context, input CrearColeccionInput) (*entity.ColeccionRegistro, error)
    ListarColecciones(ctx context.Context, limite, offset int) (*ListarColeccionesOutput, error)
    ActualizarColeccion(ctx context.Context, input ActualizarColeccionInput) (*ActualizarColeccionOutput, error)
    EliminarColeccion(ctx context.Context, input EliminarColeccionInput) (*EliminarColeccionOutput, error)
}