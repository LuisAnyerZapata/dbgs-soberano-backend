package entity

// HealthCheckResultado representa el estado de un componente de infraestructura
type HealthCheckResultado struct {
    Estado   string // "UP" o "DOWN"
    Componente string
    LatenciaMs float64
}

// SistemaHealth representa el estado general del motor DBGS
type SistemaHealth struct {
    Estado       string // "SERVING" o "DOWN"
    UptimeSegundos int64
    BaseDeDatos  HealthCheckResultado
}

// SistemaVersion contiene los metadatos de compilación del binario
type SistemaVersion struct {
    Version    string
    GitCommit  string
    BuildDate  string
    GoVersion  string
    Environment string
    Engine     string
}