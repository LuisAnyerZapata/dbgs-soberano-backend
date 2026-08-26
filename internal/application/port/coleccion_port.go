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

type RenombrarColumnaInput struct {
    NombreActual string `json:"nombre_actual"`
    NombreNuevo  string `json:"nombre_nuevo"`
}

type CambiarTipoColumnaInput struct {
    Nombre    string `json:"nombre"`
    NuevoTipo string `json:"nuevo_tipo"`
}

type ActualizarColeccionInput struct {
    Nombre          string                   `json:"nombre"`
    NuevoNombre     string                   `json:"nuevo_nombre"`
    Descripcion     string                   `json:"descripcion"`
    CamposAgregar   []entity.CampoDinamico   `json:"campos_agregar"`
    CamposRenombrar []RenombrarColumnaInput  `json:"campos_renombrar"`
    CamposTipo      []CambiarTipoColumnaInput `json:"campos_tipo"`
    CamposEliminar  []string                 `json:"campos_eliminar"`
    Confirmar       bool                     `json:"confirmar"`
}

type ActualizarColeccionOutput struct {
    ID                string `json:"id"`
    NombreLogico      string `json:"nombre_logico"`
    CamposAgregados   int    `json:"campos_agregados"`
    CamposRenombrados int    `json:"campos_renombrados"`
    CamposTipoCambiado int   `json:"campos_tipo_cambiado"`
    CamposEliminados  int    `json:"campos_eliminar"`
    NombreTablaAnterior string `json:"nombre_tabla_anterior"`
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