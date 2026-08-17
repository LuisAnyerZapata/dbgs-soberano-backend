package postgres

import (
    "context"
    "database/sql"
    "errors"
    "log"

    "DBGS_SOBERANO_BACKEND/internal/domain/entity"
    "DBGS_SOBERANO_BACKEND/internal/domain/repository"
)

type integracionPostgresRepository struct {
    db *sql.DB
}

func NewIntegracionPostgresRepository(db *sql.DB) repository.IntegracionRepository {
    return &integracionPostgresRepository{db: db}
}

// GuardarCliente persiste el cliente. Nota: Guarda el TokenHash.
func (r *integracionPostgresRepository) GuardarCliente(ctx context.Context, cliente *entity.ClienteIntegracion) error {
    // Si en el futuro usas la tabla separada de tokens_api, ajusta esta consulta
    query := `
        INSERT INTO dbgs_schema.clientes_api (id, nombre_cliente, tipo_cliente, token_hash, estatus, created_at, created_by)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
    `
    
    // Para evitar errores de NULL, generamos ID si no viene
    if cliente.ID == "" {
        cliente.ID = "00000000-0000-0000-0000-000000000000" // Se sobrescribe con el DEFAULT de BD si lo prefieres, pero por ahora enviamos dummy
    }

    _, err := r.db.ExecContext(ctx, query,
        cliente.ID,
        cliente.Nombre,
        cliente.Tipo,
        cliente.TokenHash, // Se almacena el hash, NUNCA el plano
        cliente.Estado,
        cliente.CreatedAt,
        "system", // TODO: Extraer de ctx.Value("user") como hicimos en catálogos
    )
    if err != nil {
        log.Printf("ERROR EN BD (Integracion.GuardarCliente): %v", err)
        return entity.ErrErrorInterno
    }
    return nil
}

// ValidarCredenciales busca por el hash directamente. Es la consulta más segura.
func (r *integracionPostgresRepository) ValidarCredenciales(ctx context.Context, tokenHash string) (*entity.ClienteIntegracion, error) {
    query := `
        SELECT id, nombre_cliente, tipo_cliente, version_contrato, estatus
        FROM dbgs_schema.clientes_api
        WHERE token_hash = $1 AND estatus = true
        LIMIT 1
    `
    row := r.db.QueryRowContext(ctx, query, tokenHash)

    var cliente entity.ClienteIntegracion
    err := row.Scan(&cliente.ID, &cliente.Nombre, &cliente.Tipo, &cliente.VersionContrato, &cliente.Estado)

    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, entity.ErrAccesoNoAutorizado
        }
        log.Printf("ERROR EN BD (Integracion.ValidarCredenciales): %v", err)
        return nil, entity.ErrErrorInterno
    }
    return &cliente, nil
}

func (r *integracionPostgresRepository) ObtenerClientePorID(ctx context.Context, id string) (*entity.ClienteIntegracion, error) {
    // Implementación estándar de obtención por ID (útil para la UI de administración)
    return nil, nil // TODO: Implementar cuando se haga la UI
}

func (r *integracionPostgresRepository) GuardarSolicitud(ctx context.Context, solicitud *entity.SolicitudIntegracion) error {
    // TODO: Implementar el guardado del log de traza en la tabla de solicitudes
    return nil
}