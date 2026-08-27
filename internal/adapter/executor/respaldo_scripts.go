// Package executor implementa adaptadores de salida que ejecutan procesos del
// sistema operativo (scripts bash, binarios de PostgreSQL) en nombre del dominio.
package executor

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

const backupPrefijo = "dbgs_backup_"

// MotorScripts ejecuta los respaldos PostgreSQL. En sistemas Unix delega en los
// scripts oficiales de db/backup (bash); en Windows invoca pg_dump/pg_restore
// directamente, ya que ese SO no expone bash de forma nativa. Serializa las
// operaciones con un mutex para evitar que dos pg_dump concurrentes escriban el
// mismo archivo.
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

// EjecutarCreacion genera un .dump de la base de datos y retorna la ruta del
// archivo recién creado. Usa bash+script en Unix y pg_dump directo en Windows.
func (m *MotorScripts) EjecutarCreacion(ctx context.Context, destinoDir string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if runtime.GOOS == "windows" {
		return m.ejecutarCreacionWindows(ctx, destinoDir)
	}
	return m.ejecutarCreacionUnix(ctx, destinoDir)
}

// EjecutarRestauracion restaura un .dump sobre la base de datos.
func (m *MotorScripts) EjecutarRestauracion(ctx context.Context, rutaArchivo string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if runtime.GOOS == "windows" {
		return m.ejecutarRestauracionWindows(ctx, rutaArchivo)
	}
	script := filepath.Join(m.scriptsDir, "restore_dbgs.sh")
	cmd := exec.CommandContext(ctx, "bash", script, rutaArchivo)
	cmd.Env = append(os.Environ(), append(m.entorno(), "AUTO_APPROVE=1")...)

	salida, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("pg_restore falló sobre %s: %s", rutaArchivo, truncarSalida(salida, err))
	}
	return nil
}

// ---------------------------------------------------------------------------
// Unix (bash + scripts oficiales de db/backup)
// ---------------------------------------------------------------------------

func (m *MotorScripts) ejecutarCreacionUnix(ctx context.Context, destinoDir string) (string, error) {
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

// ---------------------------------------------------------------------------
// Windows (pg_dump / pg_restore directos, sin dependencia de bash)
// ---------------------------------------------------------------------------

// ejecutarCreacionWindows replica el flujo de backup_dbgs.sh llamando al binario
// pg_dump.exe que PostgreSQL instala en el PATH de Windows.
func (m *MotorScripts) ejecutarCreacionWindows(ctx context.Context, destinoDir string) (string, error) {
	if err := os.MkdirAll(destinoDir, 0o755); err != nil {
		return "", fmt.Errorf("no se pudo crear el directorio de respaldo: %w", err)
	}

	backupFile := filepath.Join(destinoDir, fmt.Sprintf("%s%s.dump", backupPrefijo, time.Now().Format("20060102_150405")))

	cmd := exec.CommandContext(ctx, "pg_dump",
		"-h", m.dbHost,
		"-p", m.dbPort,
		"-U", m.dbUser,
		"-F", "c",
		"-b",
		"-v",
		"-f", backupFile,
		m.dbName,
	)
	cmd.Env = append(os.Environ(), "PGPASSWORD="+m.dbPassword)

	salida, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("pg_dump falló: %s", truncarSalida(salida, err))
	}

	ruta, err := archivoMasReciente(destinoDir)
	if err != nil {
		return "", fmt.Errorf("pg_dump terminó pero no se encontró el archivo de respaldo: %w", err)
	}
	return ruta, nil
}

// ejecutarRestauracionWindows replica el flujo de restore_dbgs.sh llamando al
// binario pg_restore.exe que PostgreSQL instala en el PATH de Windows.
func (m *MotorScripts) ejecutarRestauracionWindows(ctx context.Context, rutaArchivo string) error {
	if _, err := os.Stat(rutaArchivo); err != nil {
		return fmt.Errorf("el archivo de respaldo no existe: %s", rutaArchivo)
	}

	cmd := exec.CommandContext(ctx, "pg_restore",
		"-h", m.dbHost,
		"-p", m.dbPort,
		"-U", m.dbUser,
		"-d", m.dbName,
		"--clean",
		"--if-exists",
		"--no-owner",
		"-v",
		rutaArchivo,
	)
	cmd.Env = append(os.Environ(), "PGPASSWORD="+m.dbPassword)

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
// más reciente dentro de dir (el nombre exacto lo decide el motor con su timestamp).
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
