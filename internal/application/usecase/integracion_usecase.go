package usecase

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"DBGS_SOBERANO_BACKEND/internal/application/port"
	"DBGS_SOBERANO_BACKEND/internal/domain"
	"DBGS_SOBERANO_BACKEND/internal/domain/entity"
	"DBGS_SOBERANO_BACKEND/internal/domain/repository"
)

type integracionUseCase struct {
	integracionRepo repository.IntegracionRepository
}

func NewIntegracionUseCase(repo repository.IntegracionRepository) port.IntegracionPort {
	return &integracionUseCase{integracionRepo: repo}
}

func (u *integracionUseCase) RegistrarCliente(ctx context.Context, input port.RegistrarClienteInput) (*entity.ClienteIntegracion, error) {
	if input.Nombre == "" || input.Tipo == "" {
		return nil, domain.ErrDatosInvalidos
	}
	if input.VersionContrato == "" {
		input.VersionContrato = "v1"
	}

	cliente := &entity.ClienteIntegracion{
		Nombre:          input.Nombre,
		Tipo:            input.Tipo,
		VersionContrato: input.VersionContrato,
		Scopes:          input.Scopes,
		Estado:          true,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	cliente.Token = u.generarToken(cliente)

	if err := u.integracionRepo.GuardarCliente(ctx, cliente); err != nil {
		return nil, err
	}
	return cliente, nil
}

func (u *integracionUseCase) ValidarAcceso(ctx context.Context, input port.ValidarAccesoIntegracionInput) (*entity.ClienteIntegracion, error) {
	if input.ClienteID == "" || input.Token == "" {
		return nil, domain.ErrDatosInvalidos
	}

	cliente, err := u.integracionRepo.ObtenerClientePorID(ctx, input.ClienteID)
	if err != nil {
		return nil, err
	}
	if !cliente.Estado || cliente.Token != input.Token {
		return nil, domain.ErrAccesoNoAutorizado
	}
	if input.VersionContrato != "" && cliente.VersionContrato != input.VersionContrato {
		return nil, domain.ErrAccesoNoAutorizado
	}
	return cliente, nil
}

func (u *integracionUseCase) RegistrarSolicitud(ctx context.Context, input port.RegistrarSolicitudInput) error {
	if input.ClienteID == "" || input.Metodo == "" || input.Recurso == "" {
		return domain.ErrDatosInvalidos
	}
	Solicitud := &entity.SolicitudIntegracion{
		ClienteID: input.ClienteID,
		Metodo:    input.Metodo,
		Recurso:   input.Recurso,
		Estado:    input.Estado,
		Fecha:     time.Now(),
		Detalles:  input.Detalles,
	}
	return u.integracionRepo.GuardarSolicitud(ctx, Solicitud)
}

func (u *integracionUseCase) generarToken(cliente *entity.ClienteIntegracion) string {
	sum := sha256.Sum256([]byte(fmt.Sprintf("%s:%s:%d", cliente.Nombre, cliente.Tipo, time.Now().UnixNano())))
	return hex.EncodeToString(sum[:])
}
