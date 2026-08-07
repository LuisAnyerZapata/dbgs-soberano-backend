package postgres

import (
	"context"
	"database/sql"
	"time"

	"DBGS_SOBERANO_BACKEND/internal/domain/entity"
	"DBGS_SOBERANO_BACKEND/internal/domain/repository"
)

type auditoriaPostgresRepository struct {
	db *sql.DB
}

func NewAuditoriaPostgresRepository(db *sql.DB) repository.AuditoriaRepository {
	return &auditoriaPostgresRepository{db: db}
}

func (r *auditoriaPostgresRepository) RegistrarEvento(ctx context.Context, evento *entity.AuditoriaEvento) error {
	query := `
		INSERT INTO dbgs_schema.auditoria_eventos (id, usuario_id, username, operacion, recurso, detalles, resultado, ip_origen, fecha_creacion)
		VALUES (COALESCE(NULLIF($1, ''), uuid_generate_v4()), $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
	`
	row := r.db.QueryRowContext(ctx, query,
		evento.ID, evento.UsuarioID, evento.Username, evento.Operacion,
		evento.Recurso, evento.Detalles, evento.Resultado, evento.IPOrigen, evento.FechaCreacion,
	)
	var id string
	if err := row.Scan(&id); err != nil {
		return entity.ErrErrorInterno
	}
	evento.ID = id
	return nil
}

func (r *auditoriaPostgresRepository) ListarEventos(ctx context.Context, usuarioID, operacion, resultado string, fechaInicio, fechaFin *time.Time, limite, offset int) ([]entity.AuditoriaEvento, int64, error) {
	countQuery := `
		SELECT COUNT(*) FROM dbgs_schema.auditoria_eventos
		WHERE ($1 = '' OR usuario_id = $1)
		  AND ($2 = '' OR operacion = $2)
		  AND ($3 = '' OR resultado = $3)
		  AND ($4 IS NULL OR fecha_creacion >= $4)
		  AND ($5 IS NULL OR fecha_creacion <= $5)
	`
	var total int64
	if err := r.db.QueryRowContext(ctx, countQuery, usuarioID, operacion, resultado, fechaInicio, fechaFin).Scan(&total); err != nil {
		return nil, 0, entity.ErrErrorInterno
	}

	query := `
		SELECT id, usuario_id, username, operacion, recurso, detalles, resultado, ip_origen, fecha_creacion
		FROM dbgs_schema.auditoria_eventos
		WHERE ($1 = '' OR usuario_id = $1)
		  AND ($2 = '' OR operacion = $2)
		  AND ($3 = '' OR resultado = $3)
		  AND ($4 IS NULL OR fecha_creacion >= $4)
		  AND ($5 IS NULL OR fecha_creacion <= $5)
		ORDER BY fecha_creacion DESC
		LIMIT $6 OFFSET $7
	`
	rows, err := r.db.QueryContext(ctx, query, usuarioID, operacion, resultado, fechaInicio, fechaFin, limite, offset)
	if err != nil {
		return nil, 0, entity.ErrErrorInterno
	}
	defer rows.Close()

	var eventos []entity.AuditoriaEvento
	for rows.Next() {
		var e entity.AuditoriaEvento
		if err := rows.Scan(&e.ID, &e.UsuarioID, &e.Username, &e.Operacion, &e.Recurso, &e.Detalles, &e.Resultado, &e.IPOrigen, &e.FechaCreacion); err != nil {
			return nil, 0, entity.ErrErrorInterno
		}
		eventos = append(eventos, e)
	}
	return eventos, total, nil
}