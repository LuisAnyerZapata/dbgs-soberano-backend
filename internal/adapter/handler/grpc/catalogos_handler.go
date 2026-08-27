package grpc

import (
	"context"
	"strconv"

	pb "DBGS_SOBERANO_BACKEND/api/proto/v1"
	"DBGS_SOBERANO_BACKEND/internal/application/port"
	"DBGS_SOBERANO_BACKEND/internal/domain"
	"DBGS_SOBERANO_BACKEND/internal/domain/entity"

	"google.golang.org/protobuf/types/known/timestamppb"
)

type CatalogoHandler struct {
	pb.UnimplementedCatalogosServiceServer
	useCase port.CatalogoUseCasePort
}

func NewCatalogoHandler(useCase port.CatalogoUseCasePort) *CatalogoHandler {
	return &CatalogoHandler{useCase: useCase}
}

func (h *CatalogoHandler) ListarCatalogos(ctx context.Context, req *pb.ListarCatalogosRequest) (*pb.ListarCatalogosResponse, error) {
	offset := 0
	if req.GetPageToken() != "" {
		parsed, err := strconv.Atoi(req.GetPageToken())
		if err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	input := port.ObtenerCatalogosInput{
		SoloActivos: req.GetSoloActivos(),
		Limite:      int(req.GetPageSize()),
		Offset:      offset,
	}

	output, err := h.useCase.ListarCatalogos(ctx, input)
	if err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}

	var pbCatalogos []*pb.Catalogo
	for _, cat := range output.Catalogos {
		pbCatalogos = append(pbCatalogos, toCatalogoProto(cat))
	}

	nextPageToken := ""
	if len(pbCatalogos) > 0 && int64(offset)+int64(len(pbCatalogos)) < output.Total {
		nextPageToken = strconv.Itoa(offset + len(pbCatalogos))
	}

	return &pb.ListarCatalogosResponse{
		Catalogos:     pbCatalogos,
		NextPageToken: nextPageToken,
		TotalCount:    int32(output.Total),
	}, nil
}

func (h *CatalogoHandler) CrearCatalogo(ctx context.Context, req *pb.CrearCatalogoRequest) (*pb.CatalogoResponse, error) {
	errs := domain.NewValidator()
	errs.Add(domain.ValidateCodigo("codigo", req.GetCodigo()))
	errs.Add(domain.ValidateName("nombre", req.GetNombre()))
	if err := errs.Validate(); err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}

	catalogo := &entity.Catalogo{
		Codigo:      req.GetCodigo(),
		Nombre:      req.GetNombre(),
		Descripcion: req.GetDescripcion(),
	}

	created, err := h.useCase.CrearCatalogo(ctx, catalogo)
	if err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}

	return &pb.CatalogoResponse{Catalogo: toCatalogoProto(*created)}, nil
}

func (h *CatalogoHandler) ObtenerCatalogoPorID(ctx context.Context, req *pb.ObtenerCatalogoPorIDRequest) (*pb.CatalogoResponse, error) {
	if err := domain.ValidateUUID("id", req.GetId()); err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}

	catalogo, err := h.useCase.ObtenerPorID(ctx, req.GetId())
	if err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}

	return &pb.CatalogoResponse{Catalogo: toCatalogoProto(*catalogo)}, nil
}

func (h *CatalogoHandler) ActualizarCatalogo(ctx context.Context, req *pb.ActualizarCatalogoRequest) (*pb.CatalogoResponse, error) {
	errs := domain.NewValidator()
	errs.Add(domain.ValidateUUID("id", req.GetId()))
	errs.Add(domain.ValidateName("nombre", req.GetNombre()))
	if err := errs.Validate(); err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}

	catalogo := &entity.Catalogo{
		ID:          req.GetId(),
		Codigo:      "",
		Nombre:      req.GetNombre(),
		Descripcion: req.GetDescripcion(),
		Estado:      req.GetEstado(),
	}

	updated, err := h.useCase.ActualizarCatalogo(ctx, catalogo)
	if err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}

	return &pb.CatalogoResponse{Catalogo: toCatalogoProto(*updated)}, nil
}

func (h *CatalogoHandler) EliminarCatalogo(ctx context.Context, req *pb.EliminarCatalogoRequest) (*pb.EliminarCatalogoResponse, error) {
	if err := domain.ValidateUUID("id", req.GetId()); err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}

	if err := h.useCase.EliminarCatalogo(ctx, req.GetId(), "system"); err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}

	return &pb.EliminarCatalogoResponse{Exitoso: true, Mensaje: "Catálogo eliminado correctamente"}, nil
}

func toCatalogoProto(c entity.Catalogo) *pb.Catalogo {
	return &pb.Catalogo{
		Id:          c.ID,
		Codigo:      c.Codigo,
		Nombre:      c.Nombre,
		Descripcion: c.Descripcion,
		Estado:      c.Estado,
		CreatedBy:   c.CreatedBy,
		CreatedAt:   timestamppb.New(c.CreatedAt),
		UpdatedAt:   timestamppb.New(c.UpdatedAt),
	}
}
