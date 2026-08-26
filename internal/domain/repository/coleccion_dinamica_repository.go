package repository

import (
    "context"
    "DBGS_SOBERANO_BACKEND/internal/domain/entity"
)

// ColeccionDinamicaRepository define el contrato para ejecutar DDL y guardar metadatos
type ColeccionDinamicaRepository interface {
    // EjecutarDDL ejecuta una sentencia CREATE TABLE o CREATE TRIGGER de forma aislada
    EjecutarDDL(ctx context.Context, ddl string) error
    
    // GuardarMetadatos inserta el registro en el diccionario de colecciones dinámicas
    GuardarMetadatos(ctx context.Context, registro *entity.ColeccionRegistro) error

    // ObtenerMetadatosPorNombre localiza una colección activa por su nombre lógico
    ObtenerMetadatosPorNombre(ctx context.Context, nombreLogico string) (*entity.ColeccionRegistro, error)

    ListarMetadatos(ctx context.Context, limite, offset int) ([]entity.ColeccionRegistro, int64, error)

    // ActualizarMetadatos sobrescribe la estructura JSON de una colección existente
    ActualizarMetadatos(ctx context.Context, nombreLogico string, estructura []byte) error

    // DesactivarMetadatos marca una colección como inactiva (soft delete)
    DesactivarMetadatos(ctx context.Context, nombreLogico string) error

    // RenombrarMetadatos actualiza el nombre lógico y físico de una colección
    RenombrarMetadatos(ctx context.Context, nombreActual, nombreNuevo, fisicoNuevo string) error

    // EliminarMetadatos elimina el registro del diccionario de datos
    EliminarMetadatos(ctx context.Context, nombreLogico string) error
}