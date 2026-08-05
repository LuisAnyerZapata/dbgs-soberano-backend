package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// Config agrupa toda la configuración de la aplicación.
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Security SecurityConfig `mapstructure:"security"`
}

// ServerConfig contiene la configuración del servidor gRPC.
type ServerConfig struct {
	Host             string `mapstructure:"host"`
	Port             int    `mapstructure:"port"`
	EnableReflection bool   `mapstructure:"enable_reflection"`
}

// DatabaseConfig contiene los datos de conexión a PostgreSQL.
type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Name     string `mapstructure:"name"`
	SSLMode  string `mapstructure:"ssl_mode"`
	MaxConns int32  `mapstructure:"max_conns"`
	MinConns int32  `mapstructure:"min_conns"`
}

// SecurityConfig contiene parámetros clave para JWT/Auth y cifrado.
type SecurityConfig struct {
	JWTSecret      string `mapstructure:"jwt_secret"`
	TokenTTLMinutes int    `mapstructure:"token_ttl_minutes"`
}

// LoadConfig lee la configuración desde un archivo config.json o variables de entorno.
func LoadConfig(path string) (*Config, error) {
	v := viper.New()

	// 1. Configuración de archivo JSON
	for _, candidatePath := range configSearchPaths(path) {
		v.AddConfigPath(candidatePath)
	}
	v.SetConfigName("config")
	v.SetConfigType("json")

	// 2. Mapeo de Variables de Entorno (DBGS_SERVER_PORT, DBGS_DATABASE_HOST, etc.)
	v.SetEnvPrefix("DBGS")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// 3. Valores Por Defecto (Fallback)
	setDefaults(v)

	// Intenta leer el archivo de configuración si existe
	if err := v.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			return nil, fmt.Errorf("error al leer archivo de configuración: %w", err)
		}
		// Si no existe el archivo config.json, continuará usando las variables de entorno o defaults
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("error al deserializar la configuración: %w", err)
	}

	// Validaciones básicas de campos obligatorios
	if err := validateConfig(&cfg); err != nil {
		return nil, fmt.Errorf("configuración inválida: %w", err)
	}

	return &cfg, nil
}

func configSearchPaths(path string) []string {
	candidates := []string{}
	seen := map[string]struct{}{}

	addCandidate := func(candidate string) {
		if candidate == "" {
			return
		}
		cleaned := filepath.Clean(candidate)
		if _, ok := seen[cleaned]; ok {
			return
		}
		seen[cleaned] = struct{}{}
		candidates = append(candidates, cleaned)
	}

	addCandidate(path)
	addCandidate(filepath.Join(path, "config"))

	if wd, err := os.Getwd(); err == nil {
		addCandidate(wd)
		addCandidate(filepath.Join(wd, "config"))
		for _, parent := range []string{filepath.Dir(wd), filepath.Dir(filepath.Dir(wd))} {
			if parent != "." && parent != string(filepath.Separator) {
				addCandidate(parent)
				addCandidate(filepath.Join(parent, "config"))
			}
		}
	}

	if execPath, err := os.Executable(); err == nil {
		execDir := filepath.Dir(execPath)
		addCandidate(execDir)
		addCandidate(filepath.Join(execDir, "config"))
		for _, parent := range []string{filepath.Dir(execDir), filepath.Dir(filepath.Dir(execDir))} {
			if parent != "." && parent != string(filepath.Separator) {
				addCandidate(parent)
				addCandidate(filepath.Join(parent, "config"))
			}
		}
	}

	return candidates
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 50051)
	v.SetDefault("server.enable_reflection", true)

	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", 5432)
	v.SetDefault("database.user", "postgres")
	v.SetDefault("database.password", "postgres")
	v.SetDefault("database.name", "dbgs_soberano")
	v.SetDefault("database.ssl_mode", "disable")
	v.SetDefault("database.max_conns", 10)
	v.SetDefault("database.min_conns", 2)

	v.SetDefault("security.jwt_secret", "CambiarSecretPorUnoSeguroEnProduccion")
	v.SetDefault("security.token_ttl_minutes", 480)
}

func validateConfig(cfg *Config) error {
	if cfg.Server.Port <= 0 {
		return errors.New("el puerto del servidor debe ser un número positivo")
	}
	if cfg.Database.Host == "" {
		return errors.New("el host de la base de datos no puede estar vacío")
	}
	if cfg.Database.Name == "" {
		return errors.New("el nombre de la base de datos no puede estar vacío")
	}
	return nil
}