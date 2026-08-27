package postgres

import (
	"context"
	"database/sql"
	"errors"
	"log"

	"DBGS_SOBERANO_BACKEND/internal/domain/entity"
	"DBGS_SOBERANO_BACKEND/internal/domain/repository"
)

type conexionPostgresRepository struct {
	db *sql.DB
}

// scanner abstrae *sql.Row y *sql.Rows para reutilizar la lógica de escaneo.
type scanner interface {
	Scan(dest ...any) error
}

func NewConexionPostgresRepository(db *sql.DB) repository.ConexionRepository {
	return &conexionPostgresRepository{db: db}
}

func scanConexion(row scanner) (*entity.Conexion, error) {
	var c entity.Conexion
	if err := row.Scan(&c.ID, &c.Name, &c.Engine, &c.Host, &c.Port, &c.User, &c.PasswordHash, &c.Database, &c.SSLMode, &c.ReadOnly, &c.CreatedAt, &c.CreatedBy); err != nil {
		return nil, err
	}
	return &c, nil
}

const conexionCols = `id, nombre, motor, host, puerto, usuario, password_cifrado, base_datos, ssl_mode, solo_lectura, created_at, created_by`

func (r *conexionPostgresRepository) Guardar(ctx context.Context, c *entity.Conexion) error {
	query := `INSERT INTO dbgs_schema.conexiones (id, nombre, motor, host, puerto, usuario, password_cifrado, base_datos, ssl_mode, solo_lectura, created_at, created_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`
	if _, err := r.db.ExecContext(ctx, query, c.ID, c.Name, c.Engine, c.Host, c.Port, c.User, c.PasswordHash, c.Database, c.SSLMode, c.ReadOnly, c.CreatedAt, c.CreatedBy); err != nil {
		log.Printf("conexionPostgresRepository.Guardar exec error: %v", err)
		return entity.ErrErrorInterno
	}
	return nil
}

func (r *conexionPostgresRepository) ObtenerPorID(ctx context.Context, id string) (*entity.Conexion, error) {
	query := `SELECT ` + conexionCols + ` FROM dbgs_schema.conexiones WHERE id = $1`
	row := r.db.QueryRowContext(ctx, query, id)
	c, err := scanConexion(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, entity.ErrEntidadNoEncontrada
		}
		log.Printf("conexionPostgresRepository.ObtenerPorID scan error: %v", err)
		return nil, entity.ErrErrorInterno
	}
	return c, nil
}

func (r *conexionPostgresRepository) Listar(ctx context.Context, limite, offset int) ([]entity.Conexion, int64, error) {
	var total int64
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM dbgs_schema.conexiones`).Scan(&total); err != nil {
		log.Printf("conexionPostgresRepository.Listar count error: %v", err)
		return nil, 0, entity.ErrErrorInterno
	}
	query := `SELECT ` + conexionCols + ` FROM dbgs_schema.conexiones ORDER BY created_at DESC LIMIT $1 OFFSET $2`
	rows, err := r.db.QueryContext(ctx, query, limite, offset)
	if err != nil {
		log.Printf("conexionPostgresRepository.Listar query error: %v", err)
		return nil, 0, entity.ErrErrorInterno
	}
	defer rows.Close()
	var list []entity.Conexion
	for rows.Next() {
		c, err := scanConexion(rows)
		if err != nil {
			log.Printf("conexionPostgresRepository.Listar scan error: %v", err)
			return nil, 0, entity.ErrErrorInterno
		}
		list = append(list, *c)
	}
	return list, total, nil
}

func (r *conexionPostgresRepository) Actualizar(ctx context.Context, c *entity.Conexion) (*entity.Conexion, error) {
	query := `UPDATE dbgs_schema.conexiones SET nombre=$1, motor=$2, host=$3, puerto=$4, usuario=$5, password_cifrado=$6, base_datos=$7, ssl_mode=$8, solo_lectura=$9
		WHERE id=$10 RETURNING ` + conexionCols
	row := r.db.QueryRowContext(ctx, query, c.Name, c.Engine, c.Host, c.Port, c.User, c.PasswordHash, c.Database, c.SSLMode, c.ReadOnly, c.ID)
	updated, err := scanConexion(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, entity.ErrEntidadNoEncontrada
		}
		log.Printf("conexionPostgresRepository.Actualizar scan error: %v", err)
		return nil, entity.ErrErrorInterno
	}
	return updated, nil
}

func (r *conexionPostgresRepository) Eliminar(ctx context.Context, id string) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM dbgs_schema.conexiones WHERE id=$1`, id); err != nil {
		log.Printf("conexionPostgresRepository.Eliminar exec error: %v", err)
		return entity.ErrErrorInterno
	}
	return nil
}
