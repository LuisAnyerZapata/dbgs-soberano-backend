package usecase

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"DBGS_SOBERANO_BACKEND/internal/application/port"
	"DBGS_SOBERANO_BACKEND/internal/domain"
	"DBGS_SOBERANO_BACKEND/internal/domain/entity"
	"DBGS_SOBERANO_BACKEND/internal/domain/repository"

	"github.com/google/uuid"
)

// Permisos exigidos por el dominio (sembrados en db/seeds/seed_users.sql).
const (
	permisoRespaldoEjecutar     = "respaldo:ejecutar"
	permisoRestauracionEjecutar = "restauracion:ejecutar"
)

const (
	retencionDefaultDias = 30
	maximoBackupsDefault = 10
	listaLimiteDefault   = 20
	listaLimiteMaximo    = 100
	maximoDescargaBytes  = int64(1) << 30 // 1 GiB: la descarga viaja íntegra en memoria
	timeoutMotorDefecto  = 30 * time.Minute
)

// RespaldoConfig agrupa los parámetros de operación del dominio de respaldos.
type RespaldoConfig struct {
	DumpsDir         string        // Directorio físico donde viven los .dump
	TimeoutEjecucion time.Duration // Presupuesto de tiempo para pg_dump/pg_restore (0 = 30min)
}

type respaldoUseCase struct {
	repo      repository.RespaldoRepository
	seguridad repository.SeguridadRepository
	motor     port.MotorRespaldo
	cfg       RespaldoConfig
}

func NewRespaldoUseCase(repo repository.RespaldoRepository, seguridad repository.SeguridadRepository, motor port.MotorRespaldo, cfg RespaldoConfig) port.RespaldoPort {
	if cfg.TimeoutEjecucion <= 0 {
		cfg.TimeoutEjecucion = timeoutMotorDefecto
	}
	return &respaldoUseCase{repo: repo, seguridad: seguridad, motor: motor, cfg: cfg}
}

// =========================================================================================================
// CREACIÓN ASÍNCRONA DE RESPALDOS
// =========================================================================================================

// CrearRespaldo registra la operación en EN_PROGRESO y lanza pg_dump en una goroutine.
// La petición retorna de inmediato; el desenlace se consulta vía ObtenerRespaldo.
func (u *respaldoUseCase) CrearRespaldo(ctx context.Context, input port.CrearRespaldoInput) (*port.CrearRespaldoOutput, error) {
	usuario, err := u.autorizar(ctx, permisoRespaldoEjecutar)
	if err != nil {
		return nil, err
	}

	if input.Tipo == "" {
		input.Tipo = "FULL"
	}
	if input.RetencionDias <= 0 {
		input.RetencionDias = retencionDefaultDias
	}
	if input.Usuario == "" {
		input.Usuario = usuario.Username
	}

	registro := &entity.RespaldoOperacion{
		ID:             uuid.New().String(),
		Tipo:           input.Tipo,
		Estado:         entity.EstadoEnProgreso,
		Detalles:       input.Detalles,
		FechaCreacion:  time.Now(),
		RetencionDias:  input.RetencionDias,
		UsuarioCreador: input.Usuario,
	}
	if err := u.repo.GuardarRespaldo(ctx, registro); err != nil {
		return nil, err
	}

	go u.ejecutarCreacion(registro.ID, input.Detalles)

	return &port.CrearRespaldoOutput{Respaldo: registro}, nil
}

// ejecutarCreacion corre fuera del ciclo de la petición: usa un contexto propio
// con presupuesto de tiempo porque el ctx de gRPC se cancela al retornar el handler.
func (u *respaldoUseCase) ejecutarCreacion(id, detallesPrevios string) {
	defer u.recuperarPanico("CrearRespaldo", id)

	bgCtx, cancel := context.WithTimeout(context.Background(), u.cfg.TimeoutEjecucion)
	defer cancel()

	ruta, err := u.motor.EjecutarCreacion(bgCtx, u.cfg.DumpsDir)
	respaldo, errConsulta := u.repo.ObtenerRespaldo(bgCtx, id)
	if errConsulta != nil {
		log.Printf("ERROR (Respaldo.ejecutarCreacion): no se pudo recargar el registro %s: %v", id, errConsulta)
		return
	}
	respaldo.FechaFinalizacion = time.Now()

	if err != nil {
		respaldo.Estado = entity.EstadoFallido
		respaldo.Detalles = unirDetalles(detallesPrevios, err.Error())
		u.finalizar(respaldo, "ERROR")
		return
	}

	info, statErr := os.Stat(ruta)
	if statErr == nil {
		respaldo.TamanoBytes = info.Size()
	}
	respaldo.Estado = entity.EstadoCompletado
	respaldo.RutaArchivo = ruta
	respaldo.Detalles = unirDetalles(detallesPrevios, "pg_dump finalizado correctamente")
	u.finalizar(respaldo, "INFO")
}

// ObtenerRespaldo expone el estado actual para que el panel haga polling.
func (u *respaldoUseCase) ObtenerRespaldo(ctx context.Context, id string) (*port.ObtenerRespaldoOutput, error) {
	if _, err := u.autorizar(ctx, permisoRespaldoEjecutar); err != nil {
		return nil, err
	}
	if id == "" {
		return nil, domain.ErrDatosInvalidos
	}
	respaldo, err := u.repo.ObtenerRespaldo(ctx, id)
	if err != nil {
		return nil, err
	}
	return &port.ObtenerRespaldoOutput{Respaldo: respaldo}, nil
}

func (u *respaldoUseCase) ListarRespaldos(ctx context.Context, limite int) (*port.ListarRespaldosOutput, error) {
	if _, err := u.autorizar(ctx, permisoRespaldoEjecutar); err != nil {
		return nil, err
	}
	if limite <= 0 {
		limite = listaLimiteDefault
	}
	if limite > listaLimiteMaximo {
		limite = listaLimiteMaximo
	}
	respaldos, err := u.repo.ListarRespaldos(ctx, limite)
	if err != nil {
		return nil, err
	}
	return &port.ListarRespaldosOutput{Respaldos: respaldos, Total: int32(len(respaldos))}, nil
}

// DescargarRespaldo devuelve el contenido íntegro del .dump solicitado.
func (u *respaldoUseCase) DescargarRespaldo(ctx context.Context, input port.DescargarRespaldoInput) (*port.DescargarRespaldoOutput, error) {
	if _, err := u.autorizar(ctx, permisoRespaldoEjecutar); err != nil {
		return nil, err
	}
	if input.ID == "" {
		return nil, domain.ErrDatosInvalidos
	}

	respaldo, err := u.repo.ObtenerRespaldo(ctx, input.ID)
	if err != nil {
		return nil, err
	}
	if respaldo.Estado != entity.EstadoCompletado {
		return nil, fmt.Errorf("%w: solo puede descargarse un respaldo COMPLETADO", domain.ErrDatosInvalidos)
	}
	if err := u.validarRutaArchivo(respaldo.RutaArchivo); err != nil {
		return nil, err
	}

	info, err := os.Stat(respaldo.RutaArchivo)
	if err != nil {
		log.Printf("ERROR (Respaldo.DescargarRespaldo): archivo ausente %s: %v", respaldo.RutaArchivo, err)
		return nil, entity.ErrEntidadNoEncontrada
	}
	if info.Size() > maximoDescargaBytes {
		return nil, fmt.Errorf("%w: el respaldo supera el límite de descarga (%d bytes)", domain.ErrDatosInvalidos, maximoDescargaBytes)
	}

	contenido, err := os.ReadFile(respaldo.RutaArchivo)
	if err != nil {
		log.Printf("ERROR (Respaldo.DescargarRespaldo): lectura fallida %s: %v", respaldo.RutaArchivo, err)
		return nil, entity.ErrErrorInterno
	}

	return &port.DescargarRespaldoOutput{
		ID:            respaldo.ID,
		NombreArchivo: filepath.Base(respaldo.RutaArchivo),
		Contenido:     contenido,
	}, nil
}

// =========================================================================================================
// RESTAURACIÓN ASÍNCRONA
// =========================================================================================================

// RestaurarRespaldo relanza la base desde un .dump registrado (pg_restore --clean).
// Es destructiva y doblemente custodiada: permiso 'restauracion:ejecutar' + confirmación explícita.
func (u *respaldoUseCase) RestaurarRespaldo(ctx context.Context, input port.RestaurarBackupInput) (*port.RestaurarBackupOutput, error) {
	usuarioAutenticado, err := u.autorizar(ctx, permisoRestauracionEjecutar)
	if err != nil {
		return nil, err
	}
	if input.BackupID == "" || !input.Confirmar {
		return nil, fmt.Errorf("%w: la restauración requiere backup_id y confirmación explícita", domain.ErrDatosInvalidos)
	}

	respaldo, err := u.repo.ObtenerRespaldo(ctx, input.BackupID)
	if err != nil {
		return nil, err
	}
	if respaldo.Estado != entity.EstadoCompletado {
		return nil, fmt.Errorf("%w: solo puede restaurarse un respaldo COMPLETADO", domain.ErrDatosInvalidos)
	}
	if err := u.validarRutaArchivo(respaldo.RutaArchivo); err != nil {
		return nil, err
	}
	if _, err := os.Stat(respaldo.RutaArchivo); err != nil {
		return nil, entity.ErrEntidadNoEncontrada
	}

	usuario := input.Usuario
	if usuario == "" {
		usuario = usuarioAutenticado.Username
	}
	restauracion := &entity.Restauracion{
		ID:            uuid.New().String(),
		BackupID:      respaldo.ID,
		Usuario:       usuario,
		Estado:        entity.EstadoEnProgreso,
		Validado:      true, // Precondiciones verificadas: estado COMPLETADO + ruta confiable + archivo presente
		FechaCreacion: time.Now(),
		Observaciones: fmt.Sprintf("iniciada por %s desde %s", usuario, filepath.Base(respaldo.RutaArchivo)),
	}
	if err := u.repo.GuardarRestauracion(ctx, restauracion); err != nil {
		return nil, err
	}

	ruta := respaldo.RutaArchivo
	restauracionID := restauracion.ID
	go u.ejecutarRestauracion(restauracionID, ruta, usuario)

	return &port.RestaurarBackupOutput{Restauracion: restauracion}, nil
}

func (u *respaldoUseCase) ejecutarRestauracion(restauracionID, ruta, usuario string) {
	defer u.recuperarPanico("RestaurarRespaldo", restauracionID)

	bgCtx, cancel := context.WithTimeout(context.Background(), u.cfg.TimeoutEjecucion)
	defer cancel()

	err := u.motor.EjecutarRestauracion(bgCtx, ruta)

	actualizacion := &entity.Restauracion{ID: restauracionID}
	if err != nil {
		actualizacion.Estado = entity.EstadoFallido
		actualizacion.Observaciones = truncarObservaciones(err.Error())
	} else {
		actualizacion.Estado = entity.EstadoCompletado
		actualizacion.Observaciones = fmt.Sprintf("completada por %s", usuario)
	}
	if errPersistencia := u.repo.ActualizarRestauracion(bgCtx, actualizacion); errPersistencia != nil {
		log.Printf("ERROR (Respaldo.ejecutarRestauracion): no se pudo actualizar %s: %v", restauracionID, errPersistencia)
	}

	nivel := "INFO"
	mensaje := fmt.Sprintf("restauración %s completada", restauracionID)
	if err != nil {
		nivel = "ERROR"
		mensaje = fmt.Sprintf("restauración %s falló: %v", restauracionID, err)
	}
	_, _ = u.RegistrarLog(bgCtx, port.RegistrarLogInput{Nivel: nivel, Modulo: "respaldo", Mensaje: mensaje})
}

// =========================================================================================================
// RETENCIÓN Y OPERACIÓN
// =========================================================================================================

// AplicarRetencion borra archivos y registros vencidos (por edad y por tope máximo),
// protegiendo siempre las operaciones todavía en curso.
func (u *respaldoUseCase) AplicarRetencion(ctx context.Context, input port.RetencionInput) (*port.AplicarRetencionOutput, error) {
	if _, err := u.autorizar(ctx, permisoRespaldoEjecutar); err != nil {
		return nil, err
	}
	if input.DiasRetencion <= 0 {
		input.DiasRetencion = retencionDefaultDias
	}
	if input.MaximoBackups <= 0 {
		input.MaximoBackups = maximoBackupsDefault
	}

	respaldos, err := u.repo.ListarRespaldos(ctx, listaLimiteMaximo)
	if err != nil {
		return nil, err
	}

	corte := time.Now().AddDate(0, 0, -input.DiasRetencion)
	var sobrevivientes []entity.RespaldoOperacion
	var eliminados []string

	for _, respaldo := range respaldos {
		if esTerminal(respaldo.Estado) && respaldo.FechaCreacion.Before(corte) {
			u.eliminarRespaldo(ctx, respaldo, &eliminados)
			continue
		}
		sobrevivientes = append(sobrevivientes, respaldo)
	}

	// Tope duro de cantidad: elimina del más antiguo al más nuevo.
	exceso := len(sobrevivientes) - input.MaximoBackups
	for i := 0; i < exceso; i++ {
		candidato := sobrevivientes[i]
		if !esTerminal(candidato.Estado) {
			continue // nunca borramos una operación en curso por política de cupo
		}
		u.eliminarRespaldo(ctx, candidato, &eliminados)
	}

	return &port.AplicarRetencionOutput{IDsEliminados: eliminados}, nil
}

func (u *respaldoUseCase) eliminarRespaldo(ctx context.Context, respaldo entity.RespaldoOperacion, eliminados *[]string) {
	if respaldo.RutaArchivo != "" {
		if err := os.Remove(respaldo.RutaArchivo); err != nil && !errors.Is(err, os.ErrNotExist) {
			log.Printf("WARN (Respaldo.eliminarRespaldo): no se pudo borrar %s: %v", respaldo.RutaArchivo, err)
		}
	}
	if err := u.repo.EliminarRespaldo(ctx, respaldo.ID); err != nil {
		log.Printf("WARN (Respaldo.eliminarRespaldo): no se pudo borrar el registro %s: %v", respaldo.ID, err)
		return
	}
	*eliminados = append(*eliminados, respaldo.ID)
}

func (u *respaldoUseCase) RegistrarLog(ctx context.Context, input port.RegistrarLogInput) (*port.RegistroLogOutput, error) {
	if input.Nivel == "" || input.Modulo == "" || input.Mensaje == "" {
		return nil, domain.ErrDatosInvalidos
	}
	logOp := &entity.LogOperativo{
		ID:            uuid.New().String(),
		Nivel:         input.Nivel,
		Modulo:        input.Modulo,
		Mensaje:       input.Mensaje,
		FechaCreacion: time.Now(),
	}
	if err := u.repo.GuardarLog(ctx, logOp); err != nil {
		return nil, err
	}
	return &port.RegistroLogOutput{ID: logOp.ID}, nil
}

func (u *respaldoUseCase) RegistrarMetrica(ctx context.Context, input port.RegistrarMetricaInput) (*port.RegistroMetricaOutput, error) {
	if input.Nombre == "" {
		return nil, domain.ErrDatosInvalidos
	}
	metrica := &entity.MetricaSistema{
		ID:            uuid.New().String(),
		Nombre:        input.Nombre,
		Valor:         input.Valor,
		Unidad:        input.Unidad,
		FechaCreacion: time.Now(),
	}
	if err := u.repo.GuardarMetrica(ctx, metrica); err != nil {
		return nil, err
	}
	return &port.RegistroMetricaOutput{ID: metrica.ID}, nil
}

func (u *respaldoUseCase) EjecutarHealthCheck(ctx context.Context, input port.HealthCheckInput) (*port.HealthCheckOutput, error) {
	if input.Componente == "" {
		input.Componente = "motor_respaldos"
	}
	return &port.HealthCheckOutput{Estado: "OK", Mensaje: fmt.Sprintf("%s healthy", input.Componente)}, nil
}

// =========================================================================================================
// MÉTODOS AUXILIARES PRIVADOS
// =========================================================================================================

// autorizar valida sesión y permiso granular. Devuelve el usuario autenticado.
func (u *respaldoUseCase) autorizar(ctx context.Context, permiso string) (*entity.Usuario, error) {
	usuario, ok := ctx.Value(domain.CtxKeyUsuario).(*entity.Usuario)
	if !ok || usuario == nil {
		return nil, fmt.Errorf("%w: el dominio de respaldos exige un usuario autenticado", domain.ErrAccesoNoAutorizado)
	}
	permitido, err := u.seguridad.ValidarPermiso(ctx, usuario.RolID, permiso)
	if err != nil {
		return nil, domain.ErrErrorInterno
	}
	if !permitido {
		return nil, fmt.Errorf("%w: su rol no posee el permiso '%s'", domain.ErrAccesoNoAutorizado, permiso)
	}
	return usuario, nil
}

// validarRutaArchivo bloquea path traversal: el archivo debe vivir dentro de DumpsDir y ser .dump.
func (u *respaldoUseCase) validarRutaArchivo(ruta string) error {
	if ruta == "" {
		return domain.ErrDatosInvalidos
	}
	absoluta, err := filepath.Abs(filepath.Clean(ruta))
	if err != nil {
		return domain.ErrDatosInvalidos
	}
	base, err := filepath.Abs(u.cfg.DumpsDir)
	if err != nil {
		return domain.ErrErrorInterno
	}
	dentro := strings.HasPrefix(absoluta, base+string(filepath.Separator))
	if !dentro || absoluta == base || filepath.Ext(absoluta) != ".dump" {
		return fmt.Errorf("%w: la ruta del respaldo no pertenece al repositorio oficial de dumps", domain.ErrAccesoNoAutorizado)
	}
	return nil
}

func (u *respaldoUseCase) finalizar(respaldo *entity.RespaldoOperacion, nivel string) {
	if err := u.repo.ActualizarRespaldo(context.Background(), respaldo); err != nil {
		log.Printf("ERROR (Respaldo.finalizar): registro %s quedó inconsistente: %v", respaldo.ID, err)
	}
	_, _ = u.RegistrarLog(context.Background(), port.RegistrarLogInput{
		Nivel:  nivel,
		Modulo: "respaldo",
		Mensaje: fmt.Sprintf("respaldo %s terminó con estado %s (%d bytes)",
			respaldo.ID, respaldo.Estado, respaldo.TamanoBytes),
	})
}

// recuperarPanico evita que un pánico en las goroutines de fondo tumbe el servidor.
func (u *respaldoUseCase) recuperarPanico(operacion, id string) {
	if r := recover(); r != nil {
		log.Printf("PÁNICO RECUPERADO (Respaldo.%s id=%s): %v", operacion, id, r)
	}
}

func esTerminal(estado string) bool {
	return estado == entity.EstadoCompletado || estado == entity.EstadoFallido
}

func unirDetalles(previos, novedad string) string {
	if previos == "" {
		return novedad
	}
	return previos + " | " + novedad
}

func truncarObservaciones(texto string) string {
	if len(texto) > 500 {
		return texto[:500]
	}
	return texto
}
