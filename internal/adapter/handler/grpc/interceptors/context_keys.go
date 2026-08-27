package interceptors

// contextKey es un tipo no exportado para evitar colisiones en context.WithValue
type contextKey string

const (
	// KeyUsuario almacena el usuario autenticado en el contexto
	KeyUsuario contextKey = "usuario_autenticado"
	// KeyClienteIntegracion almacena el cliente de integración en el contexto
	KeyClienteIntegracion contextKey = "cliente_integracion"
)
