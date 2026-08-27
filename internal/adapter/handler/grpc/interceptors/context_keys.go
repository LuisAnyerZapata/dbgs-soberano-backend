package interceptors

import (
	"DBGS_SOBERANO_BACKEND/internal/domain"
)

// KeyUsuario almacena el usuario autenticado en el contexto.
// Es el MISMO valor y tipo que usa el dominio (domain.CtxKeyUsuario),
// garantizando que interceptor y casos de uso compartan la misma clave.
var KeyUsuario = domain.CtxKeyUsuario

// KeyClienteIntegracion almacena el cliente de integración (API Key) en el contexto.
var KeyClienteIntegracion = domain.CtxKeyClienteIntegracion
