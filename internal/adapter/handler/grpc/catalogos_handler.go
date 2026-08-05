package grpc

import (
	"context"

	"DBGS_SOBERANO_BACKEND/internal/application/port"
	pb "DBGS_SOBERANO_BACKEND/api/proto/v1"
)

type CatalogoHandler struct {
	pb.UnimplementedCatalogosServiceServer
	useCase port.CatalogoUseCasePort
}

func NewCatalogoHandler(useCase port.CatalogoUseCasePort) *CatalogoHandler {
	return &CatalogoHandler{
		useCase: useCase,
	}
}

func (h *CatalogoHandler) ListarCatalogos(ctx context.Context, req *pb.ListarCatalogosRequest) (*pb.ListarCatalogosResponse, error) {
	input := port.ObtenerCatalogosInput{
		SoloActivos: req.GetSoloActivos(),
		Limite:      int(req.GetPageSize()),
		Offset:      0,
	}

	output, err := h.useCase.ListarCatalogos(ctx, input)
	if err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}

	var pbCatalogos []*pb.Catalogo
	for _, cat := range output.Catalogos {
		estadoBool := cat.Estado == "ACTIVO" || cat.Estado == "true"
		pbCatalogos = append(pbCatalogos, &pb.Catalogo{
			Id:          cat.ID,
			Codigo:      cat.Codigo,
			Nombre:      cat.Nombre,
			Descripcion: "",
			Estado:      estadoBool,
		})
	}

	return &pb.ListarCatalogosResponse{
		Catalogos:  pbCatalogos,
		TotalCount: int32(output.Total),
	}, nil
}