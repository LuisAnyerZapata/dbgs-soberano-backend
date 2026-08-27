package domain

// ctxKey es un tipo no exportado para evitar colisiones en context.WithValue
type ctxKey string

const (
	// CtxKeyUsuario almacena el usuario autenticado en el contexto.
	// Lo usa el interceptor de autenticación (adapter) para inyectar al usuario
	// autenticado y lo leen los casos de uso (application) para autoría y RBAC.
	CtxKeyUsuario ctxKey = "usuario_autenticado"

	// CtxKeyClienteIntegracion almacena el cliente de integración (API Key) en el contexto.
	CtxKeyClienteIntegracion ctxKey = "cliente_integracion"
)
