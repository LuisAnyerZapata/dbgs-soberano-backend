package repository

import (
    "context"
)

// DatosDinamicosRepository define el contrato para ejecutar consultas DML genéricas y seguras
type DatosDinamicosRepository interface {
    // Listar ejecuta un SELECT dinámico sobre la tabla física aplicando paginación
    Listar(ctx context.Context, nombreFisico string, limite, offset int) ([]map[string]interface{}, int64, error)
    
    // ObtenerPorID ejecuta un SELECT dinámico filtrando por el UUID soberano
    ObtenerPorID(ctx context.Context, nombreFisico string, id string) (map[string]interface{}, error)
    
    // Insertar ejecuta un INSERT dinámico inyectando los campos del mapa y el usuario auditor
    Insertar(ctx context.Context, nombreFisico string, datos map[string]interface{}, createdBy string) (string, error)
    
    // Actualizar ejecuta un UPDATE dinámico filtrando por ID, actualizando solo los campos del mapa
    Actualizar(ctx context.Context, nombreFisico string, id string, datos map[string]interface{}, updatedBy string) (map[string]interface{}, error)
    
    // Eliminar ejecuta un DELETE dinámico filtrando por ID
    Eliminar(ctx context.Context, nombreFisico string, id string) error
}