package usecase

import (
    "fmt"
    "regexp"
    "strings"

    "DBGS_SOBERANO_BACKEND/internal/domain"
    "DBGS_SOBERANO_BACKEND/internal/domain/entity"

    "github.com/lib/pq" // Necesario para QuoteIdentifier
)

var (
    // regexNombreSeguro solo permite letras minúsculas, números y guiones bajos.
    // Obliga a empezar con letra o guion bajo. Previene SQL Injection a nivel de DDL.
    regexNombreSeguro = regexp.MustCompile(`^[a-z_][a-z0-9_]*$`)
)

const DBGS_SCHEMA = "dbgs_schema"

// GenerarSQLCreacionTabla traduce una definición de colección a un DDL seguro de PostgreSQL.
// Retorna el script SQL y un error si la validación falla.
func GenerarSQLCreacionTabla(coleccion *entity.ColeccionDefinicion) (string, error) {
    if coleccion == nil || coleccion.Nombre == "" {
        return "", domain.ErrDatosInvalidos
    }

    nombreTabla := strings.ToLower(coleccion.Nombre)
    if !regexNombreSeguro.MatchString(nombreTabla) {
        return "", fmt.Errorf("nombre de tabla inválido: solo se permiten letras minúsculas, números y guiones bajos (ej: mi_tabla_01)")
    }

    // Separamos el esquema del nombre de la tabla física
    nombreFisico := "dyn_" + nombreTabla

    var columnasSQL []string
    indicesSQL := []string{}

    // 1. Inyectar columnas soberanas obligatorias
    columnasSQL = append(columnasSQL, fmt.Sprintf("%s UUID PRIMARY KEY DEFAULT uuid_generate_v4()", pq.QuoteIdentifier("id")))
    columnasSQL = append(columnasSQL, fmt.Sprintf("%s TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP", pq.QuoteIdentifier("created_at")))
    columnasSQL = append(columnasSQL, fmt.Sprintf("%s TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP", pq.QuoteIdentifier("updated_at")))
    columnasSQL = append(columnasSQL, fmt.Sprintf("%s VARCHAR(100) NOT NULL", pq.QuoteIdentifier("created_by")))
    columnasSQL = append(columnasSQL, fmt.Sprintf("%s VARCHAR(100)", pq.QuoteIdentifier("updated_by")))

    // 2. Procesar campos definidos por el usuario
    for _, campo := range coleccion.Campos {
        nombreCampo := strings.ToLower(campo.Nombre)
        
        if nombreCampo == "id" || nombreCampo == "created_at" || nombreCampo == "updated_at" || nombreCampo == "created_by" || nombreCampo == "updated_by" {
            return "", fmt.Errorf("el nombre de campo '%s' está reservado para el sistema", nombreCampo)
        }

        if !regexNombreSeguro.MatchString(nombreCampo) {
            return "", fmt.Errorf("nombre de campo inválido: '%s'", nombreCampo)
        }

        tipoSQL, err := mapearTipoPostgres(campo.Tipo)
        if err != nil {
            return "", err
        }

        definicion := fmt.Sprintf("%s %s", pq.QuoteIdentifier(nombreCampo), tipoSQL)
        
        if !campo.Nulo {
            definicion += " NOT NULL"
        }
        if campo.Unico {
            definicion += " UNIQUE"
            // Formato correcto: esquema."tabla" 
            indicesSQL = append(indicesSQL, fmt.Sprintf("CREATE UNIQUE INDEX IF NOT EXISTS idx_%s_%s ON %s.%s (%s)", 
                nombreFisico, nombreCampo, DBGS_SCHEMA, pq.QuoteIdentifier(nombreFisico), pq.QuoteIdentifier(nombreCampo)))
        }

        columnasSQL = append(columnasSQL, definicion)
    }

    // 3. Ensamblar el DDL final (ESQUEMA SIN COMILLAS . TABLA CON COMILLAS)
    query := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s.%s (\n  %s\n);", 
        DBGS_SCHEMA, 
        pq.QuoteIdentifier(nombreFisico), 
        strings.Join(columnasSQL, ",\n  "))

    // 4. Añadir índices únicos
    for _, idx := range indicesSQL {
        query += "\n" + idx + ";"
    }

    return query, nil
}

// ObtenerNombreTablaCompleto retorna el nombre físico con el prefijo "dyn_"
func ObtenerNombreTablaCompleto(nombreLogico string) string {
    return DBGS_SCHEMA + ".dyn_" + strings.ToLower(nombreLogico)
}

// mapearTipoPostgres traduce nuestro enum seguro a tipos nativos de PostgreSQL
func mapearTipoPostgres(tipo entity.FieldType) (string, error) {
    switch tipo {
    case entity.FieldTypeString:
        return "VARCHAR(255)", nil
    case entity.FieldTypeText:
        return "TEXT", nil
    case entity.FieldTypeInt:
        return "INTEGER", nil
    case entity.FieldTypeFloat:
        return "NUMERIC(15,4)", nil // NUMERIC es más seguro que FLOAT para dinero/cantidades
    case entity.FieldTypeBoolean:
        return "BOOLEAN", nil
    case entity.FieldTypeUUID:
        return "UUID", nil
    case entity.FieldTypeJSON:
        return "JSONB", nil
    case entity.FieldTypeDate:
        return "TIMESTAMPTZ", nil
    default:
        return "", fmt.Errorf("tipo de dato no soportado para tablas dinámicas: %s", tipo)
    }
}

// GenerarSQLTriggerAuditoria crea el comando para vincular el trigger forense a la nueva tabla
func GenerarSQLTriggerAuditoria(nombreTablaFisico string) string {
    // Separamos el string "dbgs_schema.dyn_xxx" en el esquema y la tabla
    partes := strings.Split(nombreTablaFisico, ".")
    esquema := partes[0]
    tabla := partes[1]
    
    // Los nombres de trigger en Postgres son únicos por tabla
    nombreTrigger := "trg_audit_" + tabla
    
    // Se eliminó el DROP TRIGGER porque la tabla es recién creada, no hay nada que dropear.
    // Esto asegura que sea una SOLA sentencia SQL, resolviendo el problema con lib/pq.
    return fmt.Sprintf(`
        CREATE TRIGGER %s
        AFTER INSERT OR UPDATE OR DELETE ON %s.%s
        FOR EACH ROW EXECUTE FUNCTION %s.fn_auditar_cambios();
    `, 
        pq.QuoteIdentifier(nombreTrigger), 
        esquema, pq.QuoteIdentifier(tabla),
        esquema, 
    )
}