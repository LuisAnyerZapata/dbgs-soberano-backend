package port

import "context"

// --- Inputs ---

type ListarRegistrosInput struct {
    NombreTabla string
    Limite      int
    Offset      int
}

type ObtenerRegistroInput struct {
    NombreTabla string
    ID          string
}

type CrearRegistroInput struct {
    NombreTabla string
    Datos       map[string]interface{} // Viene del google.protobuf.Struct del Proto
}

type ActualizarRegistroInput struct {
    NombreTabla string
    ID          string
    Datos       map[string]interface{}
}

type EliminarRegistroInput struct {
    NombreTabla string
    ID          string
}

// --- Outputs ---

type ListarRegistrosOutput struct {
    Registros []map[string]interface{}
    Total     int64
}

// --- Puerto (Interfaz) ---

type DatosDinamicosPort interface {
    ListarRegistros(ctx context.Context, input ListarRegistrosInput) (*ListarRegistrosOutput, error)
    ObtenerRegistro(ctx context.Context, input ObtenerRegistroInput) (map[string]interface{}, error)
    CrearRegistro(ctx context.Context, input CrearRegistroInput) (string, error)
    ActualizarRegistro(ctx context.Context, input ActualizarRegistroInput) (map[string]interface{}, error)
    EliminarRegistro(ctx context.Context, input EliminarRegistroInput) error
}