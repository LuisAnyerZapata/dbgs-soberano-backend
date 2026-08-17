package usecase

import (
    "context"
    "database/sql"
    "fmt"
    "time"

    "DBGS_SOBERANO_BACKEND/internal/application/port"
    "DBGS_SOBERANO_BACKEND/internal/domain/entity"
)

// BuildInfo almacena las variables inyectadas en tiempo de compilación (ldflags)
type BuildInfo struct {
    Version    string
    GitCommit  string
    BuildDate  string
    GoVersion  string
    Environment string
}

type sistemaUseCase struct {
    db         *sql.DB
    buildInfo  BuildInfo
    startTime  time.Time
}

// NewSistemaUseCase crea un nuevo caso de uso de sistema.
// Recibe la conexión a BD y los metadatos de compilación.
func NewSistemaUseCase(db *sql.DB, info BuildInfo) port.SistemaPort {
    return &sistemaUseCase{
        db:        db,
        buildInfo: info,
        startTime: time.Now(),
    }
}

func (uc *sistemaUseCase) ObtenerHealth(ctx context.Context) (*entity.SistemaHealth, error) {
    health := &entity.SistemaHealth{
        Estado:       "SERVING",
        UptimeSegundos: int64(time.Since(uc.startTime).Seconds()),
        BaseDeDatos: entity.HealthCheckResultado{
            Estado:    "DOWN",
            Componente: "PostgreSQL",
        },
    }

    // Contexto con timeout estricto para no colgar el endpoint
    pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
    defer cancel()

    start := time.Now()
    err := uc.db.PingContext(pingCtx)
    latency := time.Since(start).Seconds() * 1000 // Conversión a milisegundos

    if err == nil {
        health.BaseDeDatos.Estado = "UP"
        health.BaseDeDatos.LatenciaMs = latency
    } else {
        health.Estado = "DOWN"
        health.BaseDeDatos.LatenciaMs = latency
    }

    return health, nil
}

func (uc *sistemaUseCase) ObtenerVersion(ctx context.Context) (*entity.SistemaVersion, error) {
    return &entity.SistemaVersion{
        Version:    uc.buildInfo.Version,
        GitCommit:  uc.buildInfo.GitCommit,
        BuildDate:  uc.buildInfo.BuildDate,
        GoVersion:  uc.buildInfo.GoVersion,
        Environment: uc.buildInfo.Environment,
        Engine:     fmt.Sprintf("DBGS Core Engine v%s", uc.buildInfo.Version),
    }, nil
}