package grpc

import (
	"context"
	"fmt"
	"time"

	pb "DBGS_SOBERANO_BACKEND/api/proto/v1"
	"DBGS_SOBERANO_BACKEND/internal/application/port"
	"DBGS_SOBERANO_BACKEND/internal/domain"
	"DBGS_SOBERANO_BACKEND/internal/domain/entity"
)

type ColeccionesHandler struct {
	pb.UnimplementedColeccionesServiceServer
	useCase port.ColeccionPort
}

func NewColeccionesHandler(uc port.ColeccionPort) *ColeccionesHandler {
	return &ColeccionesHandler{useCase: uc}
}

func (h *ColeccionesHandler) CrearColeccion(ctx context.Context, req *pb.CrearColeccionRequest) (*pb.CrearColeccionResponse, error) {
	errs := domain.NewValidator()
	errs.Add(domain.ValidateName("nombre", req.GetNombre()))
	if len(req.GetCampos()) == 0 {
		errs.Add(domain.RequiredError("campos"))
	}
	for i, c := range req.GetCampos() {
		errs.Add(domain.ValidateRequired(fmt.Sprintf("campos[%d].nombre", i), c.GetNombre()))
		errs.Add(domain.ValidateRequired(fmt.Sprintf("campos[%d].tipo", i), c.GetTipo()))
	}
	if err := errs.Validate(); err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}

	var camposDominio []entity.CampoDinamico
	for _, c := range req.Campos {
		camposDominio = append(camposDominio, entity.CampoDinamico{
			Nombre:      c.GetNombre(),
			Tipo:        entity.FieldType(c.GetTipo()),
			Nulo:        c.GetNulo(),
			Unico:       c.GetUnico(),
			Descripcion: c.GetDescripcion(),
		})
	}

	input := port.CrearColeccionInput{
		Nombre:        req.GetNombre(),
		Descripcion:   req.GetDescripcion(),
		Campos:        camposDominio,
		InstitucionID: req.GetInstitucionId(),
	}

	registro, err := h.useCase.CrearColeccion(ctx, input)
	if err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}

	return &pb.CrearColeccionResponse{
		Id:           registro.ID,
		NombreLogico: registro.NombreLogico,
		NombreFisico: registro.NombreFisico,
		Mensaje:      "Colección creada y tabla dinámica generada exitosamente",
	}, nil
}

func (h *ColeccionesHandler) ListarColecciones(ctx context.Context, req *pb.ListarColeccionesRequest) (*pb.ListarColeccionesResponse, error) {
	output, err := h.useCase.ListarColecciones(ctx, int(req.GetLimit()), int(req.GetOffset()))
	if err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}

	var pbColecciones []*pb.ColeccionRegistroProto
	for _, c := range output.Colecciones {
		pbColecciones = append(pbColecciones, &pb.ColeccionRegistroProto{
			Id:             c.ID,
			NombreLogico:   c.NombreLogico,
			NombreFisico:   c.NombreFisico,
			Descripcion:    c.Descripcion,
			EstructuraJson: string(c.EstructuraJSON),
			EstaActiva:     c.EstaActiva,
			CreatedAt:      c.CreatedAt.Format(time.RFC3339),
		})
	}

	return &pb.ListarColeccionesResponse{
		Colecciones: pbColecciones,
		Total:       int32(output.Total),
	}, nil
}

func (h *ColeccionesHandler) ActualizarColeccion(ctx context.Context, req *pb.ActualizarColeccionRequest) (*pb.ActualizarColeccionResponse, error) {
	if err := domain.ValidateRequired("nombre", req.GetNombre()); err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}

	var camposAgregar []entity.CampoDinamico
	for _, c := range req.CamposAgregar {
		camposAgregar = append(camposAgregar, entity.CampoDinamico{
			Nombre:      c.GetNombre(),
			Tipo:        entity.FieldType(c.GetTipo()),
			Nulo:        c.GetNulo(),
			Unico:       c.GetUnico(),
			Descripcion: c.GetDescripcion(),
		})
	}

	var camposRenombrar []port.RenombrarColumnaInput
	for _, r := range req.CamposRenombrar {
		camposRenombrar = append(camposRenombrar, port.RenombrarColumnaInput{
			NombreActual: r.GetNombreActual(),
			NombreNuevo:  r.GetNombreNuevo(),
		})
	}

	var camposTipo []port.CambiarTipoColumnaInput
	for _, t := range req.CamposTipo {
		camposTipo = append(camposTipo, port.CambiarTipoColumnaInput{
			Nombre:    t.GetNombre(),
			NuevoTipo: t.GetNuevoTipo(),
		})
	}

	input := port.ActualizarColeccionInput{
		Nombre:          req.GetNombre(),
		NuevoNombre:     req.GetNuevoNombre(),
		Descripcion:     req.GetDescripcion(),
		CamposAgregar:   camposAgregar,
		CamposRenombrar: camposRenombrar,
		CamposTipo:      camposTipo,
		CamposEliminar:  req.CamposEliminar,
		Confirmar:       req.GetConfirmar(),
	}

	output, err := h.useCase.ActualizarColeccion(ctx, input)
	if err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}

	var msg string
	if output.NombreTablaAnterior != "" {
		msg = fmt.Sprintf("Tabla renombrada desde %s. Operaciones aplicadas: %d agregados, %d renombrados, %d tipo cambiado, %d eliminados",
			output.NombreTablaAnterior, output.CamposAgregados, output.CamposRenombrados, output.CamposTipoCambiado, output.CamposEliminados)
	} else {
		msg = fmt.Sprintf("Operaciones aplicadas: %d agregados, %d renombrados, %d tipo cambiado, %d eliminados",
			output.CamposAgregados, output.CamposRenombrados, output.CamposTipoCambiado, output.CamposEliminados)
	}

	return &pb.ActualizarColeccionResponse{
		Id:                  output.ID,
		NombreLogico:        output.NombreLogico,
		CamposAgregados:     int32(output.CamposAgregados),
		CamposRenombrados:   int32(output.CamposRenombrados),
		CamposTipoCambiado:  int32(output.CamposTipoCambiado),
		CamposEliminar:      int32(output.CamposEliminados),
		NombreTablaAnterior: output.NombreTablaAnterior,
		Mensaje:             msg,
	}, nil
}

func (h *ColeccionesHandler) EliminarColeccion(ctx context.Context, req *pb.EliminarColeccionRequest) (*pb.EliminarColeccionResponse, error) {
	if err := domain.ValidateRequired("nombre", req.GetNombre()); err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}

	input := port.EliminarColeccionInput{
		Nombre:    req.GetNombre(),
		Confirmar: req.GetConfirmar(),
	}

	output, err := h.useCase.EliminarColeccion(ctx, input)
	if err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}

	var mensaje string
	if output.AccionRealizada == "eliminada" {
		mensaje = "Colección eliminada permanentemente"
	} else {
		mensaje = "Colección desactivada (tabla física preservada)"
	}

	return &pb.EliminarColeccionResponse{
		NombreLogico:    output.NombreLogico,
		AccionRealizada: output.AccionRealizada,
		Mensaje:         mensaje,
	}, nil
}
