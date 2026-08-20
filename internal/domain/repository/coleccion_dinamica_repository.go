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
    ListarMetadatos(ctx context.Context, limite, offset int) ([]entity.ColeccionRegistro, int64, error)
}