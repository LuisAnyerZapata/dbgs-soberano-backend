package postgres

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"time"

	"DBGS_SOBERANO_BACKEND/internal/domain/entity"
	"DBGS_SOBERANO_BACKEND/internal/domain/repository"
)

type respaldoPostgresRepository struct {
	db *sql.DB
}

func NewRespaldoPostgresRepository(db *sql.DB) repository.RespaldoRepository {
	return &respaldoPostgresRepository{db: db}
}

const columnasRespaldo = `id, tipo, estado, ruta_archivo, tamano_bytes, detalles,
	fecha_creacion, fecha_finalizacion, retencion_dias, usuario_creador`

func escanearRespaldo(row interface{ Scan(dest ...any) error }) (*entity.RespaldoOperacion, error) {
	var r entity.RespaldoOperacion
	var finalizacion sql.NullTime
	if err := row.Scan(
		&r.ID, &r.Tipo, &r.Estado, &r.RutaArchivo, &r.TamanoBytes, &r.Detalles,
		&r.FechaCreacion, &finalizacion, &r.RetencionDias, &r.UsuarioCreador,
	); err != nil {
		return nil, err
	}
	r.FechaFinalizacion = finalizacion.Time
	return &r, nil
}

func (r *respaldoPostgresRepository) GuardarRespaldo(ctx context.Context, respaldo *entity.RespaldoOperacion) error {
	query := `
		INSERT INTO dbgs_schema.respaldos
			(id, tipo, estado, ruta_archivo, tamano_bytes, detalles, fecha_creacion, fecha_finalizacion, retencion_dias, usuario_creador)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
	_, err := r.db.ExecContext(ctx, query,
		respaldo.ID, respaldo.Tipo, respaldo.Estado, respaldo.RutaArchivo, respaldo.TamanoBytes,
		respaldo.Detalles, respaldo.FechaCreacion, nuloSiCero(respaldo.FechaFinalizacion),
		respaldo.RetencionDias, respaldo.UsuarioCreador,
	)
	if err != nil {
		log.Printf("ERROR EN BD (Respaldo.GuardarRespaldo id=%s): %v", respaldo.ID, err)
		return entity.ErrErrorInterno
	}
	return nil
}

func (r *respaldoPostgresRepository) ObtenerRespaldo(ctx context.Context, id string) (*entity.RespaldoOperacion, error) {
	query := `SELECT ` + columnasRespaldo + ` FROM dbgs_schema.respaldos WHERE id = $1`
	respaldo, err := escanearRespaldo(r.db.QueryRowContext(ctx, query, id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, entity.ErrEntidadNoEncontrada
		}
		log.Printf("ERROR EN BD (Respaldo.ObtenerRespaldo id=%s): %v", id, err)
		return nil, entity.ErrErrorInterno
	}
	return respaldo, nil
}

func (r *respaldoPostgresRepository) ListarRespaldos(ctx context.Context, limite int) ([]entity.RespaldoOperacion, error) {
	query := `SELECT ` + columnasRespaldo + `
		FROM dbgs_schema.respaldos ORDER BY fecha_creacion DESC LIMIT $1`
	rows, err := r.db.QueryContext(ctx, query, limite)
	if err != nil {
		log.Printf("ERROR EN BD (Respaldo.ListarRespaldos): %v", err)
		return nil, entity.ErrErrorInterno
	}
	defer rows.Close()

	var resultado []entity.RespaldoOperacion
	for rows.Next() {
		respaldo, err := escanearRespaldo(rows)
		if err != nil {
			log.Printf("ERROR EN BD (Respaldo.ListarRespaldos scan): %v", err)
			return nil, entity.ErrErrorInterno
		}
		resultado = append(resultado, *respaldo)
	}
	return resultado, rows.Err()
}

func (r *respaldoPostgresRepository) ActualizarRespaldo(ctx context.Context, respaldo *entity.RespaldoOperacion) error {
	query := `
		UPDATE dbgs_schema.respaldos
		SET estado = $2, ruta_archivo = $3, tamano_bytes = $4, detalles = $5,
		    fecha_finalizacion = $6, retencion_dias = $7
		WHERE id = $1`
	res, err := r.db.ExecContext(ctx, query,
		respaldo.ID, respaldo.Estado, respaldo.RutaArchivo, respaldo.TamanoBytes,
		respaldo.Detalles, nuloSiCero(respaldo.FechaFinalizacion), respaldo.RetencionDias,
	)
	if err != nil {
		log.Printf("ERROR EN BD (Respaldo.ActualizarRespaldo id=%s): %v", respaldo.ID, err)
		return entity.ErrErrorInterno
	}
	if filas, _ := res.RowsAffected(); filas == 0 {
		return entity.ErrEntidadNoEncontrada
	}
	return nil
}

func (r *respaldoPostgresRepository) EliminarRespaldo(ctx context.Context, id string) error {
	query := `DELETE FROM dbgs_schema.respaldos WHERE id = $1`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		log.Printf("ERROR EN BD (Respaldo.EliminarRespaldo id=%s): %v", id, err)
		return entity.ErrErrorInterno
	}
	if filas, _ := res.RowsAffected(); filas == 0 {
		return entity.ErrEntidadNoEncontrada
	}
	return nil
}

func (r *respaldoPostgresRepository) GuardarRestauracion(ctx context.Context, restauracion *entity.Restauracion) error {
	query := `
		INSERT INTO dbgs_schema.restauraciones (id, backup_id, usuario, estado, validado, observaciones, fecha_creacion)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := r.db.ExecContext(ctx, query,
		restauracion.ID, restauracion.BackupID, restauracion.Usuario, restauracion.Estado,
		restauracion.Validado, restauracion.Observaciones, restauracion.FechaCreacion,
	)
	if err != nil {
		log.Printf("ERROR EN BD (Respaldo.GuardarRestauracion id=%s): %v", restauracion.ID, err)
		return entity.ErrErrorInterno
	}
	return nil
}

// ActualizarRestauracion persiste el desenlace de la operación asíncrona.
func (r *respaldoPostgresRepository) ActualizarRestauracion(ctx context.Context, restauracion *entity.Restauracion) error {
	query := `
		UPDATE dbgs_schema.restauraciones
		SET estado = $2, validado = $3, observaciones = $4
		WHERE id = $1`
	res, err := r.db.ExecContext(ctx, query,
		restauracion.ID, restauracion.Estado, restauracion.Validado, restauracion.Observaciones,
	)
	if err != nil {
		log.Printf("ERROR EN BD (Respaldo.ActualizarRestauracion id=%s): %v", restauracion.ID, err)
		return entity.ErrErrorInterno
	}
	if filas, _ := res.RowsAffected(); filas == 0 {
		return entity.ErrEntidadNoEncontrada
	}
	return nil
}

func (r *respaldoPostgresRepository) GuardarLog(ctx context.Context, logOp *entity.LogOperativo) error {
	query := `
		INSERT INTO dbgs_schema.logs_operativos (id, nivel, modulo, mensaje, fecha_creacion)
		VALUES ($1, $2, $3, $4, $5)`
	_, err := r.db.ExecContext(ctx, query,
		logOp.ID, logOp.Nivel, logOp.Modulo, logOp.Mensaje, logOp.FechaCreacion,
	)
	if err != nil {
		log.Printf("ERROR EN BD (Respaldo.GuardarLog): %v", err)
		return entity.ErrErrorInterno
	}
	return nil
}

func (r *respaldoPostgresRepository) GuardarMetrica(ctx context.Context, metrica *entity.MetricaSistema) error {
	query := `
		INSERT INTO dbgs_schema.metricas_sistema (id, nombre, valor, unidad, fecha_creacion)
		VALUES ($1, $2, $3, $4, $5)`
	_, err := r.db.ExecContext(ctx, query,
		metrica.ID, metrica.Nombre, metrica.Valor, metrica.Unidad, metrica.FechaCreacion,
	)
	if err != nil {
		log.Printf("ERROR EN BD (Respaldo.GuardarMetrica): %v", err)
		return entity.ErrErrorInterno
	}
	return nil
}

// nuloSiCero traduce la fecha cero de Go a NULL de PostgreSQL.
func nuloSiCero(t time.Time) any {
	if t.IsZero() {
		return nil
	}
	return t
}

var _ repository.RespaldoRepository = (*respaldoPostgresRepository)(nil)
