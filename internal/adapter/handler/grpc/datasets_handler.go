package grpc

import (
	"context"

	dbgsv1 "DBGS_SOBERANO_BACKEND/api/proto/v1"
	"DBGS_SOBERANO_BACKEND/internal/application/port"
	"DBGS_SOBERANO_BACKEND/internal/domain"
	"DBGS_SOBERANO_BACKEND/internal/domain/entity"

	"google.golang.org/protobuf/types/known/timestamppb"
)

type DatasetsHandler struct {
	dbgsv1.UnimplementedDatasetsServiceServer
	useCase port.DatasetPort
}

func NewDatasetsHandler(uc port.DatasetPort) *DatasetsHandler {
	return &DatasetsHandler{useCase: uc}
}

func (h *DatasetsHandler) ListarDatasets(ctx context.Context, req *dbgsv1.ListarDatasetsRequest) (*dbgsv1.ListarDatasetsResponse, error) {
	clasificacion := ""
	if req.GetFiltroClasificacion() != dbgsv1.ClasificacionSeguridad_CLASIFICACION_UNSPECIFIED {
		clasificacion = req.GetFiltroClasificacion().String()
	}

	input := port.ObtenerDatasetsInput{
		Clasificacion: clasificacion,
		Limite:        int(req.GetLimit()),
		Offset:        int(req.GetOffset()),
	}

	output, err := h.useCase.ListarDatasets(ctx, input)
	if err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}

	var pbDatasets []*dbgsv1.Dataset
	for _, ds := range output.Datasets {
		pbDatasets = append(pbDatasets, toDatasetProto(ds))
	}

	return &dbgsv1.ListarDatasetsResponse{
		Datasets:     pbDatasets,
		TotalRecords: int32(output.Total),
	}, nil
}

func (h *DatasetsHandler) ObtenerDatasetPorID(ctx context.Context, req *dbgsv1.ObtenerDatasetPorIDRequest) (*dbgsv1.DatasetResponse, error) {
	if err := domain.ValidateUUID("id", req.GetId()); err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}

	dataset, err := h.useCase.ObtenerDatasetPorID(ctx, req.GetId())
	if err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}

	return &dbgsv1.DatasetResponse{Dataset: toDatasetProto(*dataset)}, nil
}

func (h *DatasetsHandler) CrearDataset(ctx context.Context, req *dbgsv1.CrearDatasetRequest) (*dbgsv1.DatasetResponse, error) {
	errs := domain.NewValidator()
	errs.Add(domain.ValidateRequired("nombre", req.GetNombre()))
	errs.Add(domain.ValidateRequired("fuente_dato_id", req.GetFuenteDatoId()))
	if err := errs.Validate(); err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}

	dataset := &entity.ConjuntoDato{
		FuenteDatoID:    req.GetFuenteDatoId(),
		Nombre:          req.GetNombre(),
		Proposito:       req.GetProposito(),
		PropietarioDato: req.GetPropietarioDato(),
		Clasificacion:   req.GetClasificacion().String(),
	}

	created, err := h.useCase.CrearDataset(ctx, dataset)
	if err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}

	return &dbgsv1.DatasetResponse{Dataset: toDatasetProto(*created)}, nil
}

func (h *DatasetsHandler) ActualizarDataset(ctx context.Context, req *dbgsv1.ActualizarDatasetRequest) (*dbgsv1.DatasetResponse, error) {
	errs := domain.NewValidator()
	errs.Add(domain.ValidateUUID("id", req.GetId()))
	errs.Add(domain.ValidateRequired("nombre", req.GetNombre()))
	if err := errs.Validate(); err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}

	dataset := &entity.ConjuntoDato{
		ID:            req.GetId(),
		Nombre:        req.GetNombre(),
		Proposito:     req.GetProposito(),
		Clasificacion: req.GetClasificacion().String(),
		Estado:        req.GetEstado(),
	}

	updated, err := h.useCase.ActualizarDataset(ctx, dataset)
	if err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}

	return &dbgsv1.DatasetResponse{Dataset: toDatasetProto(*updated)}, nil
}

func toDatasetProto(ds entity.ConjuntoDato) *dbgsv1.Dataset {
	return &dbgsv1.Dataset{
		Id:              ds.ID,
		FuenteDatoId:    ds.FuenteDatoID,
		Nombre:          ds.Nombre,
		Proposito:       ds.Proposito,
		PropietarioDato: ds.PropietarioDato,
		Clasificacion:   toProtoClasificacion(ds.Clasificacion),
		Estado:          ds.Estado,
		CreatedBy:       ds.CreatedBy,
		CreatedAt:       timestamppb.New(ds.CreatedAt),
		UpdatedAt:       timestamppb.New(ds.UpdatedAt),
	}
}

func toProtoClasificacion(value string) dbgsv1.ClasificacionSeguridad {
	switch value {
	case "PUBLICO":
		return dbgsv1.ClasificacionSeguridad_PUBLICO
	case "RESTRINGIDO":
		return dbgsv1.ClasificacionSeguridad_RESTRINGIDO
	case "CONFIDENCIAL":
		return dbgsv1.ClasificacionSeguridad_CONFIDENCIAL
	case "SECRETO":
		return dbgsv1.ClasificacionSeguridad_SECRETO
	default:
		return dbgsv1.ClasificacionSeguridad_CLASIFICACION_UNSPECIFIED
	}
}
