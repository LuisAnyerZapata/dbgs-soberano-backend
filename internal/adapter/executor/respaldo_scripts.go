// Package executor implementa adaptadores de salida que ejecutan procesos del
// sistema operativo (scripts bash, binarios de PostgreSQL) en nombre del dominio.
package executor

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
)

// MotorScripts ejecuta los scripts oficiales de db/backup (pg_dump/pg_restore)
// inyectando las credenciales por entorno. Serializa las operaciones con un
// mutex para evitar que dos pg_dump concurrentes escriban el mismo archivo.
type MotorScripts struct {
	scriptsDir string
	dbHost     string
	dbPort     string
	dbUser     string
	dbPassword string
	dbName     string

	mu sync.Mutex
}

// NewMotorScripts construye el motor inyectando la ruta de los scripts y los
// parámetros de conexión leídos de la configuración central.
func NewMotorScripts(scriptsDir, host string, port int, user, password, name string) *MotorScripts {
	return &MotorScripts{
		scriptsDir: scriptsDir,
		dbHost:     host,
		dbPort:     fmt.Sprintf("%d", port),
		dbUser:     user,
		dbPassword: password,
		dbName:     name,
	}
}

// EjecutarCreacion lanza backup_dbgs.sh contra destinoDir y retorna la ruta
// del .dump recién creado.
func (m *MotorScripts) EjecutarCreacion(ctx context.Context, destinoDir string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	script := filepath.Join(m.scriptsDir, "backup_dbgs.sh")
	cmd := exec.CommandContext(ctx, "bash", script, destinoDir)
	cmd.Env = append(os.Environ(), m.entorno()...)

	salida, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("pg_dump falló: %s", truncarSalida(salida, err))
	}

	ruta, err := archivoMasReciente(destinoDir)
	if err != nil {
		return "", fmt.Errorf("el script terminó pero no se encontró el archivo de respaldo: %w", err)
	}
	return ruta, nil
}

// EjecutarRestauracion lanza restore_dbgs.sh con AUTO_APPROVE=1 sobre rutaArchivo.
func (m *MotorScripts) EjecutarRestauracion(ctx context.Context, rutaArchivo string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	script := filepath.Join(m.scriptsDir, "restore_dbgs.sh")
	cmd := exec.CommandContext(ctx, "bash", script, rutaArchivo)
	cmd.Env = append(os.Environ(), append(m.entorno(), "AUTO_APPROVE=1")...)

	salida, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("pg_restore falló sobre %s: %s", rutaArchivo, truncarSalida(salida, err))
	}
	return nil
}

// entorno propaga las credenciales como variables DB_*; el script les da
// precedencia sobre su lectura de config.json.
func (m *MotorScripts) entorno() []string {
	return []string{
		"DB_HOST=" + m.dbHost,
		"DB_PORT=" + m.dbPort,
		"DB_USER=" + m.dbUser,
		"DB_PASSWORD=" + m.dbPassword,
		"DB_NAME=" + m.dbName,
	}
}

// archivoMasReciente localiza el dbgs_backup_*.dump con fecha de modificación
// más reciente dentro de dir (el nombre exacto lo decide el script con su timestamp).
func archivoMasReciente(dir string) (string, error) {
	entradas, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}
	var mejor string
	var mejorMod time.Time
	for _, entrada := range entradas {
		if entrada.IsDir() {
			continue
		}
		nombre := entrada.Name()
		if filepath.Ext(nombre) != ".dump" {
			continue
		}
		info, err := entrada.Info()
		if err != nil {
			continue
		}
		if info.ModTime().After(mejorMod) {
			mejorMod = info.ModTime()
			mejor = filepath.Join(dir, nombre)
		}
	}
	if mejor == "" {
		return "", fmt.Errorf("no hay archivos .dump en %s", dir)
	}
	return mejor, nil
}

func truncarSalida(salida []byte, err error) string {
	detalle := string(salida)
	if len(detalle) > 800 {
		detalle = detalle[len(detalle)-800:] // conservamos el final, donde está el error real
	}
	if err != nil {
		return detalle + " | " + err.Error()
	}
	return detalle
}
