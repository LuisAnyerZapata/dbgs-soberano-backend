package usecase

import (
	"context"
	"fmt"
	"time"

	"DBGS_SOBERANO_BACKEND/internal/application/port"
	"DBGS_SOBERANO_BACKEND/internal/domain"
	"DBGS_SOBERANO_BACKEND/internal/domain/entity"
	"DBGS_SOBERANO_BACKEND/internal/domain/repository"
)

type respaldoUseCase struct {
	respaldoRepo repository.RespaldoRepository
}

func NewRespaldoUseCase(repo repository.RespaldoRepository) port.RespaldoPort {
	return &respaldoUseCase{respaldoRepo: repo}
}

func (u *respaldoUseCase) CrearRespaldo(ctx context.Context, input port.CrearRespaldoInput) (*port.CrearRespaldoOutput, error) {
	if input.Tipo == "" || input.Ruta == "" {
		return nil, domain.ErrDatosInvalidos
	}

	respaldo := &entity.RespaldoOperacion{
		ID:             fmt.Sprintf("backup-%d", time.Now().UnixNano()),
		Tipo:           input.Tipo,
		Estado:         "COMPLETADO",
		RutaArchivo:    input.Ruta,
		Detalles:       input.Detalles,
		FechaCreacion:  time.Now(),
		RetencionDias:  input.Retencion,
		UsuarioCreador: input.Usuario,
	}
	if respaldo.RetencionDias <= 0 {
		respaldo.RetencionDias = 7
	}
	if err := u.respaldoRepo.GuardarRespaldo(ctx, respaldo); err != nil {
		return nil, err
	}

	return &port.CrearRespaldoOutput{ID: respaldo.ID, Estado: respaldo.Estado}, nil
}

func (u *respaldoUseCase) ValidarRestauracion(ctx context.Context, input port.RestaurarBackupInput) (*port.RestaurarBackupOutput, error) {
	if input.BackupID == "" {
		return nil, domain.ErrDatosInvalidos
	}
	backup, err := u.respaldoRepo.ObtenerRespaldo(ctx, input.BackupID)
	if err != nil {
		return nil, err
	}
	if backup.Estado != "COMPLETADO" {
		return nil, domain.ErrDatosInvalidos
	}

	restauracion := &entity.Restauracion{
		ID:            fmt.Sprintf("restore-%d", time.Now().UnixNano()),
		BackupID:      backup.ID,
		Usuario:       input.Usuario,
		Estado:        "VALIDADO",
		Validado:      true,
		FechaCreacion: time.Now(),
		Observaciones: "restauración validada por el sistema",
	}
	if err := u.respaldoRepo.GuardarRestauracion(ctx, restauracion); err != nil {
		return nil, err
	}

	return &port.RestaurarBackupOutput{Validado: true, Estado: restauracion.Estado}, nil
}

func (u *respaldoUseCase) AplicarRetencion(ctx context.Context, input port.RetencionInput) ([]string, error) {
	if input.DiasRetencion <= 0 {
		input.DiasRetencion = 30
	}
	if input.MaximoBackups <= 0 {
		input.MaximoBackups = 5
	}

	backups, err := u.respaldoRepo.ListarRespaldos(ctx, input.MaximoBackups)
	if err != nil {
		return nil, err
	}

	var removed []string
	cutoff := time.Now().AddDate(0, 0, -input.DiasRetencion)
	for _, backup := range backups {
		if backup.FechaCreacion.Before(cutoff) {
			_ = u.respaldoRepo.EliminarRespaldo(ctx, backup.ID)
			removed = append(removed, backup.ID)
		}
	}
	return removed, nil
}

func (u *respaldoUseCase) RegistrarLog(ctx context.Context, input port.RegistrarLogInput) (*port.RegistroLogOutput, error) {
	if input.Nivel == "" || input.Modulo == "" || input.Mensaje == "" {
		return nil, domain.ErrDatosInvalidos
	}
	log := &entity.LogOperativo{
		ID:            fmt.Sprintf("log-%d", time.Now().UnixNano()),
		Nivel:         input.Nivel,
		Modulo:        input.Modulo,
		Mensaje:       input.Mensaje,
		FechaCreacion: time.Now(),
	}
	if err := u.respaldoRepo.GuardarLog(ctx, log); err != nil {
		return nil, err
	}
	return &port.RegistroLogOutput{ID: log.ID}, nil
}

func (u *respaldoUseCase) RegistrarMetrica(ctx context.Context, input port.RegistrarMetricaInput) (*port.RegistroMetricaOutput, error) {
	if input.Nombre == "" {
		return nil, domain.ErrDatosInvalidos
	}
	metrica := &entity.MetricaSistema{
		ID:            fmt.Sprintf("metric-%d", time.Now().UnixNano()),
		Nombre:        input.Nombre,
		Valor:         input.Valor,
		Unidad:        input.Unidad,
		FechaCreacion: time.Now(),
	}
	if err := u.respaldoRepo.GuardarMetrica(ctx, metrica); err != nil {
		return nil, err
	}
	return &port.RegistroMetricaOutput{ID: metrica.ID}, nil
}

func (u *respaldoUseCase) EjecutarHealthCheck(ctx context.Context, input port.HealthCheckInput) (*port.HealthCheckOutput, error) {
	if input.Componente == "" {
		input.Componente = "system"
	}
	return &port.HealthCheckOutput{Estado: "OK", Mensaje: fmt.Sprintf("%s healthy", input.Componente)}, nil
}
